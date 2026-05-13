package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jasonlabz/gentol/embedded"
)

// 默认模板项目的模块路径（模板项目 go.mod 中的 module 名）
const DefaultTemplateModulePath = "github.com/jasonlabz/generate-example-project"

// 默认模板项目的短名称（目录名和项目名）
const DefaultTemplateProjectName = "generate-example-project"

// 默认模板仓库地址（可被 --template_repo 覆盖）
const DefaultTemplateRepoURL = "https://github.com/jasonlabz/generate-example-project.git"

// 离线缓存目录（~/.gentol/cache/）
var gentolCacheDir string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	gentolCacheDir = filepath.Join(homeDir, ".gentol", "cache")
}

// cacheMeta 缓存元数据，记录缓存来源和创建时间
type cacheMeta struct {
	SourceURL string    `json:"source_url"`
	CreatedAt time.Time `json:"created_at"`
	FileCount int       `json:"file_count"`
}

// memoryFile 表示内存中的一个文件
type memoryFile struct {
	Path    string // 相对路径（使用 / 分隔符）
	Content []byte
	Mode    fs.FileMode
}

// cloneToMemory 使用系统 git 克隆到临时目录，然后加载到内存
// 使用系统 git（而非 go-git）以复用 gitconfig 中的代理等配置
func cloneToMemory(repoURL string) (files []*memoryFile, err error) {
	log.Printf("Cloning template from repository: %s\n", repoURL)

	tmpDir, err := os.MkdirTemp("", "gentol-clone-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir failed: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone %s failed: %w", repoURL, err)
	}

	return loadDirToMemory(tmpDir)
}

// loadDirToMemory 从本地目录读取文件到内存
func loadDirToMemory(srcDir string) ([]*memoryFile, error) {
	log.Printf("Loading template from local directory: %s\n", srcDir)

	var files []*memoryFile

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 统一使用 / 作为分隔符
		relPath = filepath.ToSlash(relPath)

		// 跳过根目录自身
		if relPath == "." {
			return nil
		}

		// 跳过 .git 目录
		if relPath == ".git" || strings.HasPrefix(relPath, ".git/") {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// 跳过 go.sum
		if filepath.Base(path) == "go.sum" {
			return nil
		}

		// 跳过二进制文件
		if isBinaryFile(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files = append(files, &memoryFile{
			Path:    relPath,
			Content: content,
			Mode:    info.Mode(),
		})

		return nil
	})

	return files, err
}

// replaceInMemoryFiles 在内存中对所有文件执行替换
func replaceInMemoryFiles(files []*memoryFile, oldModulePath, newModulePath, oldProjectName, newProjectName string) {
	needsTwoPhase := (oldModulePath == oldProjectName) && (newModulePath != newProjectName)

	for _, f := range files {
		// 1. 替换文件内容
		if needsTwoPhase {
			f.Content = replaceContentTwoPhase(f.Content, f.Path, oldModulePath, newModulePath, newProjectName)
		} else {
			f.Content = replaceContentSimple(f.Content, oldModulePath, newModulePath, oldProjectName, newProjectName)
		}

		// 2. 替换文件路径中的项目名
		if oldProjectName != "" && newProjectName != "" && oldProjectName != newProjectName {
			f.Path = strings.ReplaceAll(f.Path, oldProjectName, newProjectName)
		}
	}
}

// replaceContentSimple 简单替换：模块路径和项目名不同（标准情况）
func replaceContentSimple(content []byte, oldModulePath, newModulePath, oldProjectName, newProjectName string) []byte {
	result := content

	// 先替换长的模块路径
	if oldModulePath != "" && oldModulePath != newModulePath {
		result = bytes.ReplaceAll(result, []byte(oldModulePath), []byte(newModulePath))
	}

	// 再替换短项目名（此时模块路径已被替换，不会冲突）
	if oldProjectName != "" && oldProjectName != newProjectName && oldProjectName != oldModulePath {
		result = bytes.ReplaceAll(result, []byte(oldProjectName), []byte(newProjectName))
	}

	return result
}

// replaceContentTwoPhase 两阶段替换：模块路径和项目名相同（如模板 module = "demo"）
func replaceContentTwoPhase(content []byte, filePath, oldModulePath, newModulePath, newProjectName string) []byte {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		// Go 文件：先全部替换为完整模块路径，再对非 import 行修正为短名称
		result := bytes.ReplaceAll(content, []byte(oldModulePath), []byte(newModulePath))
		result = replaceGoNonImportPaths(result, newModulePath, newProjectName)
		return result

	case ".mod":
		// go.mod：module 行保留完整路径
		return replaceGoModModulePath(content, oldModulePath, newModulePath)

	default:
		// 其他文件（yaml, Makefile, thrift, sh, ps1, md 等）：全部用短名称替换
		return bytes.ReplaceAll(content, []byte(oldModulePath), []byte(newProjectName))
	}
}

// replaceGoNonImportPaths 替换 Go 文件中非 import 行的完整模块路径为短项目名
func replaceGoNonImportPaths(content []byte, fullModulePath, shortName string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	inImportBlock := false
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)

		// 检测 import 块的开始和结束
		if bytes.HasPrefix(trimmed, []byte("import")) && bytes.Contains(trimmed, []byte("(")) {
			inImportBlock = true
			continue
		}
		if inImportBlock && bytes.Contains(trimmed, []byte(")")) {
			inImportBlock = false
			continue
		}
		if bytes.HasPrefix(trimmed, []byte("import")) {
			// 单行 import：不替换
			continue
		}

		// 在 import 块内：不替换
		if inImportBlock {
			continue
		}

		// 非 import 行：将完整模块路径替换为短名称
		if bytes.Contains(line, []byte(fullModulePath)) {
			lines[i] = bytes.ReplaceAll(line, []byte(fullModulePath), []byte(shortName))
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// replaceGoModModulePath 替换 go.mod 中的 module 行
func replaceGoModModulePath(content []byte, oldModulePath, newModulePath string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("module ")) {
			lines[i] = bytes.Replace(line, []byte(oldModulePath), []byte(newModulePath), 1)
			break
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// isBinaryFile 检测文件是否为二进制文件
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true, ".webp": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".mp3": true, ".mp4": true, ".wav": true, ".avi": true, ".mov": true,
		".sqlite": true, ".db": true,
	}
	if binaryExts[ext] {
		return true
	}

	// 对于无扩展名或未知扩展名的文件，读取前 512 字节检测
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}
	buf = buf[:n]

	return bytes.IndexByte(buf, 0) >= 0
}

// writeProjectFromMemory 将内存中的文件写入磁盘目标目录
func writeProjectFromMemory(files []*memoryFile, targetDir string) error {
	// 先收集所有需要创建的目录
	dirs := make(map[string]bool)
	for _, f := range files {
		dir := filepath.Dir(filepath.Join(targetDir, f.Path))
		dirs[dir] = true
	}

	// 创建所有目录
	for dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s failed: %w", dir, err)
		}
	}

	// 写入所有文件
	for _, f := range files {
		targetPath := filepath.Join(targetDir, f.Path)

		perm := f.Mode
		if perm == 0 {
			perm = fs.FileMode(0644)
		}

		// 脚本文件需要可执行权限
		ext := strings.ToLower(filepath.Ext(f.Path))
		if ext == ".sh" || ext == ".ps1" {
			perm = fs.FileMode(0755)
		}

		if err := os.WriteFile(targetPath, f.Content, perm); err != nil {
			return fmt.Errorf("write file %s failed: %w", targetPath, err)
		}

		log.Printf("writing %s\n", targetPath)
	}

	return nil
}

// runGoModTidy 在项目目录执行 go mod tidy
func runGoModTidy(projectDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	return nil
}

// extractProjectName 从模块路径中提取项目短名称
func extractProjectName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if len(parts[i]) > 0 {
			return parts[i]
		}
	}
	return modulePath
}

// getTemplateModuleInfo 获取模板项目的模块路径和项目名
func getTemplateModuleInfo(templateSource string, useLocalDir bool) (templateModulePath, templateProjectName string, err error) {
	if useLocalDir && templateSource != "" {
		// 从本地目录的 go.mod 读取模块路径
		modFile := filepath.Join(templateSource, "go.mod")
		if modPath, found := getModuleName(modFile); found {
			return modPath, extractProjectName(modPath), nil
		}
		// 无法读取 go.mod，使用默认值
		log.Printf("Warning: cannot read module path from %s, using default: %s\n", modFile, DefaultTemplateModulePath)
		return DefaultTemplateModulePath, DefaultTemplateProjectName, nil
	}

	// 从 git 仓库：使用默认的模板模块路径
	// （模板仓库的 go.mod 中 module 必须与 DefaultTemplateModulePath 一致）
	return DefaultTemplateModulePath, DefaultTemplateProjectName, nil
}

// cloneAndReplaceProject 从模板创建新项目（内存化流程）
// 整个流程：加载模板到内存 → 内存中替换 → 写入磁盘
// offline=true 时仅使用本地缓存，不访问网络
func cloneAndReplaceProject(newModulePath, templateSource string, useLocalDir bool, offline bool) error {
	newProjectName := extractProjectName(newModulePath)
	if newProjectName == "" {
		return fmt.Errorf("invalid project name from module path: %s", newModulePath)
	}

	// 确定模板模块路径和项目名
	templateModulePath, templateProjectName, err := getTemplateModuleInfo(templateSource, useLocalDir)
	if err != nil {
		return err
	}

	// 阶段1：加载模板到内存（不写入磁盘）
	var memFiles []*memoryFile

	if useLocalDir && templateSource != "" {
		memFiles, err = loadDirToMemory(templateSource)
	} else {
		repoURL := templateSource
		if repoURL == "" {
			repoURL = DefaultTemplateRepoURL
		}
		memFiles, err = loadTemplateWithCache(repoURL, offline)
	}
	if err != nil {
		return fmt.Errorf("load template failed: %w", err)
	}

	log.Printf("Template loaded: %d files\n", len(memFiles))

	// 阶段2：在内存中执行替换（模块路径 + 项目名 + 文件路径）
	log.Printf("Replacing module path: %s -> %s\n", templateModulePath, newModulePath)
	log.Printf("Replacing project name: %s -> %s\n", templateProjectName, newProjectName)
	replaceInMemoryFiles(memFiles, templateModulePath, newModulePath, templateProjectName, newProjectName)

	// 阶段3：写入磁盘目标目录
	targetDir := filepath.Join(".", newProjectName)
	if IsExist(targetDir) {
		return fmt.Errorf("project directory already exists: %s, please remove it and try again", targetDir)
	}

	if err := writeProjectFromMemory(memFiles, targetDir); err != nil {
		return fmt.Errorf("write project failed: %w", err)
	}

	// 阶段4：执行 go mod tidy
	log.Println("Running go mod tidy...")
	if err := runGoModTidy(targetDir); err != nil {
		log.Printf("Warning: go mod tidy failed (you may need to run it manually): %v\n", err)
	}

	return nil
}

// updateProjectFromTemplate 从模板更新已有项目（内存化流程）
// 与 new 的区别：目标目录已存在，模板文件覆盖同名文件，已有项目中的其他文件保持不变
// offline=true 时仅使用本地缓存，不访问网络
func updateProjectFromTemplate(projectDir, currentModulePath, templateSource string, useLocalDir bool, offline bool) error {
	// 确定模板模块路径和项目名
	templateModulePath, templateProjectName, err := getTemplateModuleInfo(templateSource, useLocalDir)
	if err != nil {
		return err
	}

	// 阶段1：加载模板到内存
	var memFiles []*memoryFile

	if useLocalDir && templateSource != "" {
		memFiles, err = loadDirToMemory(templateSource)
	} else {
		repoURL := templateSource
		if repoURL == "" {
			repoURL = DefaultTemplateRepoURL
		}
		memFiles, err = loadTemplateWithCache(repoURL, offline)
	}
	if err != nil {
		return fmt.Errorf("load template failed: %w", err)
	}

	log.Printf("Template loaded: %d files\n", len(memFiles))

	// 阶段2：在内存中执行替换（用当前项目的模块路径替换模板的）
	currentProjectName := extractProjectName(currentModulePath)
	log.Printf("Replacing module path: %s -> %s\n", templateModulePath, currentModulePath)
	log.Printf("Replacing project name: %s -> %s\n", templateProjectName, currentProjectName)
	replaceInMemoryFiles(memFiles, templateModulePath, currentModulePath, templateProjectName, currentProjectName)

	// 阶段3：写入已有项目目录（覆盖同名文件，不删除项目中已有但模板中没有的文件）
	if err := writeProjectFromMemory(memFiles, projectDir); err != nil {
		return fmt.Errorf("update project failed: %w", err)
	}

	// 阶段4：执行 go mod tidy
	log.Println("Running go mod tidy...")
	if err := runGoModTidy(projectDir); err != nil {
		log.Printf("Warning: go mod tidy failed (you may need to run it manually): %v\n", err)
	}

	return nil
}

// --- 缓存相关函数 ---

// cacheKey 根据仓库 URL 生成缓存键（SHA256 前16位）
func cacheKey(repoURL string) string {
	h := sha256.Sum256([]byte(repoURL))
	return fmt.Sprintf("%x", h[:8])
}

// saveTemplateCache 将内存文件缓存到本地磁盘
// 缓存格式：~/.gentol/cache/<key>/template.tar.gz + meta.json
func saveTemplateCache(repoURL string, files []*memoryFile) error {
	key := cacheKey(repoURL)
	cacheDir := filepath.Join(gentolCacheDir, key)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("create cache directory %s failed: %w", cacheDir, err)
	}

	// 写入 meta.json
	meta := &cacheMeta{
		SourceURL: repoURL,
		CreatedAt: time.Now(),
		FileCount: len(files),
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache meta failed: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "meta.json"), metaData, 0644); err != nil {
		return fmt.Errorf("write cache meta failed: %w", err)
	}

	// 写入 template.tar.gz
	tarPath := filepath.Join(cacheDir, "template.tar.gz")
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("create cache tar file failed: %w", err)
	}
	defer tarFile.Close()

	gw := gzip.NewWriter(tarFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, f := range files {
		hdr := &tar.Header{
			Name:    f.Path,
			Mode:    int64(f.Mode),
			Size:    int64(len(f.Content)),
			ModTime: time.Now(),
		}
		if hdr.Mode == 0 {
			hdr.Mode = 0644
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("write tar header for %s failed: %w", f.Path, err)
		}
		if _, err := tw.Write(f.Content); err != nil {
			return fmt.Errorf("write tar content for %s failed: %w", f.Path, err)
		}
	}

	log.Printf("Template cached to %s (%d files)\n", cacheDir, len(files))
	return nil
}

// loadTemplateCache 从本地缓存读取模板文件
// 返回 nil 表示缓存不存在或读取失败
func loadTemplateCache(repoURL string) ([]*memoryFile, error) {
	key := cacheKey(repoURL)
	cacheDir := filepath.Join(gentolCacheDir, key)
	tarPath := filepath.Join(cacheDir, "template.tar.gz")

	if !IsExist(tarPath) {
		return nil, nil // 缓存不存在
	}

	f, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("open cache file failed: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("decompress cache failed: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var files []*memoryFile
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read cache tar failed: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read cache entry %s failed: %w", hdr.Name, err)
		}

		files = append(files, &memoryFile{
			Path:    hdr.Name,
			Content: content,
			Mode:    fs.FileMode(hdr.Mode),
		})
	}

	// 读取元数据打印日志
	metaData, err := os.ReadFile(filepath.Join(cacheDir, "meta.json"))
	if err == nil {
		var meta cacheMeta
		if json.Unmarshal(metaData, &meta) == nil {
			log.Printf("Loaded template from cache (cached at %s, %d files)\n",
				meta.CreatedAt.Format("2006-01-02 15:04:05"), meta.FileCount)
		}
	}

	return files, nil
}

// parseTarGzBytes 解析 tar.gz 字节流为内存文件列表
func parseTarGzBytes(data []byte) ([]*memoryFile, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompress failed: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var files []*memoryFile

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar entry failed: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read tar content %s failed: %w", hdr.Name, err)
		}

		files = append(files, &memoryFile{
			Path:    hdr.Name,
			Content: content,
			Mode:    fs.FileMode(hdr.Mode),
		})
	}

	return files, nil
}

// loadEmbeddedTemplate 从编译时嵌入的模板数据加载
// 返回 nil 表示嵌入数据为空（placeholder）
func loadEmbeddedTemplate() ([]*memoryFile, error) {
	data := embedded.TemplateData
	if len(data) == 0 {
		return nil, nil
	}

	files, err := parseTarGzBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parse embedded template failed: %w", err)
	}

	if len(files) == 0 {
		return nil, nil // placeholder tar.gz
	}

	log.Printf("Loaded template from embedded data (%d files)\n", len(files))
	return files, nil
}

// loadTemplateWithCache 带缓存和嵌入机制的模板加载
// 加载优先级：
//   - 默认模板（在线）：远程 git → 嵌入数据 → 本地缓存
//   - 自定义模板（在线）：远程 git → 本地缓存
//   - --offline 模式：嵌入数据 → 本地缓存（跳过网络）
func loadTemplateWithCache(repoURL string, offline bool) ([]*memoryFile, error) {
	isDefaultTemplate := (repoURL == "" || repoURL == DefaultTemplateRepoURL)

	// 1. 在线模式：优先尝试远程 git clone（获取最新模板）
	var cloneErr error
	if !offline {
		memFiles, err := cloneToMemory(repoURL)
		if err == nil {
			// clone 成功，更新缓存
			if cacheErr := saveTemplateCache(repoURL, memFiles); cacheErr != nil {
				log.Printf("Warning: failed to save template cache: %v\n", cacheErr)
			}
			return memFiles, nil
		}
		cloneErr = err
		log.Printf("Remote clone failed: %v\n", cloneErr)
	}

	// 2. 对于默认模板，尝试嵌入数据（编译时内置的快照）
	if isDefaultTemplate {
		files, err := loadEmbeddedTemplate()
		if err != nil {
			log.Printf("Warning: embedded template invalid: %v\n", err)
		} else if files != nil {
			return files, nil
		}
	}

	// 3. 回退到本地缓存
	if offline {
		files, err := loadTemplateCache(repoURL)
		if err != nil {
			return nil, fmt.Errorf("offline mode: read cache failed: %w", err)
		}
		if files == nil {
			return nil, fmt.Errorf("offline mode: no cached template found for %s, please run `gentol new` with network first", repoURL)
		}
		return files, nil
	}

	// 在线模式但网络和嵌入都失败了，最后尝试本地缓存
	log.Printf("Falling back to local cache...\n")
	files, cacheErr := loadTemplateCache(repoURL)
	if cacheErr != nil {
		return nil, fmt.Errorf("remote clone failed (%v), embedded template unavailable, and cache read failed (%w)", cloneErr, cacheErr)
	}
	if files == nil {
		return nil, fmt.Errorf("remote clone failed (%v) and no cache available, please check network connection", cloneErr)
	}

	return files, nil
}

// --- 磁盘操作辅助函数 ---

// copyFromDir 从本地目录复制模板项目到目标目录（磁盘方式，用于旧模式兼容）
func copyFromDir(srcDir, targetDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 跳过 .git 目录
		if relPath == ".git" || strings.HasPrefix(relPath, ".git"+string(filepath.Separator)) {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// 注意：getParentPath 定义在 db_handler.go 中
