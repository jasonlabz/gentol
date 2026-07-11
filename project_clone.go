package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jasonlabz/gentol/embedded"
)

// 默认模板项目的模块路径（模板项目 go.mod 中的 module 名）
const DefaultTemplateModulePath = "github.com/jasonlabz/generate-example-project"

// 默认模板项目的短名称（目录名和项目名）
const DefaultTemplateProjectName = "generate-example-project"

// 默认模板仓库地址（可被 --template_repo 覆盖）
const DefaultTemplateRepoURL = "https://github.com/jasonlabz/generate-example-project.git"

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

// getModulePathFromMemory 从内存文件列表中提取 go.mod 中的模块路径
func getModulePathFromMemory(files []*memoryFile) (modulePath string, found bool) {
	for _, f := range files {
		if f.Path == "go.mod" {
			scanner := bufio.NewScanner(bytes.NewReader(f.Content))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					return strings.TrimPrefix(line, "module "), true
				}
			}
			break
		}
	}
	return "", false
}

// cloneAndReplaceProject 从模板创建新项目（内存化流程）
// 整个流程：加载模板到内存 → 从 go.mod 读取真实模块路径 → 内存中替换 → 写入磁盘
func cloneAndReplaceProject(newModulePath, templateSource string, useLocalDir bool) error {
	newProjectName := extractProjectName(newModulePath)
	if newProjectName == "" {
		return fmt.Errorf("invalid project name from module path: %s", newModulePath)
	}

	// 阶段1：加载模板到内存（不写入磁盘）
	var memFiles []*memoryFile
	var err error

	if useLocalDir && templateSource != "" {
		memFiles, err = loadDirToMemory(templateSource)
	} else {
		repoURL := templateSource
		if repoURL == "" {
			repoURL = DefaultTemplateRepoURL
		}
		memFiles, err = loadTemplate(repoURL, templateSource == "")
	}
	if err != nil {
		return fmt.Errorf("load template failed: %w", err)
	}

	log.Printf("Template loaded: %d files\n", len(memFiles))

	// 阶段2：从内存中 go.mod 读取模板的真实模块路径
	templateModulePath, found := getModulePathFromMemory(memFiles)
	if !found {
		return fmt.Errorf("cannot find go.mod in template, unable to determine template module path")
	}
	templateProjectName := extractProjectName(templateModulePath)

	log.Printf("Detected template module path: %s, project name: %s\n", templateModulePath, templateProjectName)

	// 阶段3：在内存中执行替换（模块路径 + 项目名 + 文件路径）
	log.Printf("Replacing module path: %s -> %s\n", templateModulePath, newModulePath)
	log.Printf("Replacing project name: %s -> %s\n", templateProjectName, newProjectName)
	replaceInMemoryFiles(memFiles, templateModulePath, newModulePath, templateProjectName, newProjectName)

	// 阶段4：写入磁盘目标目录
	targetDir := filepath.Join(".", newProjectName)
	if IsExist(targetDir) {
		return fmt.Errorf("project directory already exists: %s, please remove it and try again", targetDir)
	}

	if err := writeProjectFromMemory(memFiles, targetDir); err != nil {
		return fmt.Errorf("write project failed: %w", err)
	}

	// 阶段5：执行 go mod tidy
	log.Println("Running go mod tidy...")
	if err := runGoModTidy(targetDir); err != nil {
		log.Printf("Warning: go mod tidy failed (you may need to run it manually): %v\n", err)
	}

	return nil
}

// updateProjectFromTemplate 从模板更新已有项目（内存化流程）
// 与 new 的区别：目标目录已存在，模板文件覆盖同名文件，已有项目中的其他文件保持不变
func updateProjectFromTemplate(projectDir, currentModulePath, templateSource string, useLocalDir bool) error {

	// 阶段1：加载模板到内存
	var memFiles []*memoryFile
	var err error

	if useLocalDir && templateSource != "" {
		memFiles, err = loadDirToMemory(templateSource)
	} else {
		repoURL := templateSource
		if repoURL == "" {
			repoURL = DefaultTemplateRepoURL
		}
		memFiles, err = loadTemplate(repoURL, templateSource == "")
	}
	if err != nil {
		return fmt.Errorf("load template failed: %w", err)
	}

	log.Printf("Template loaded: %d files\n", len(memFiles))

	// 阶段2：从内存中 go.mod 读取模板的真实模块路径
	templateModulePath, found := getModulePathFromMemory(memFiles)
	if !found {
		return fmt.Errorf("cannot find go.mod in template, unable to determine template module path")
	}
	templateProjectName := extractProjectName(templateModulePath)

	log.Printf("Detected template module path: %s, project name: %s\n", templateModulePath, templateProjectName)

	// 阶段3：在内存中执行替换（用当前项目的模块路径替换模板的）
	currentProjectName := extractProjectName(currentModulePath)
	log.Printf("Replacing module path: %s -> %s\n", templateModulePath, currentModulePath)
	log.Printf("Replacing project name: %s -> %s\n", templateProjectName, currentProjectName)
	replaceInMemoryFiles(memFiles, templateModulePath, currentModulePath, templateProjectName, currentProjectName)

	// 阶段4：写入已有项目目录（覆盖同名文件，不删除项目中已有但模板中没有的文件）
	if err := writeProjectFromMemory(memFiles, projectDir); err != nil {
		return fmt.Errorf("update project failed: %w", err)
	}

	// 阶段5：执行 go mod tidy
	log.Println("Running go mod tidy...")
	if err := runGoModTidy(projectDir); err != nil {
		log.Printf("Warning: go mod tidy failed (you may need to run it manually): %v\n", err)
	}

	return nil
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

// loadTemplate 加载模板：远程 git clone → 嵌入数据（默认模板 fallback）
func loadTemplate(repoURL string, isDefault bool) ([]*memoryFile, error) {
	memFiles, err := cloneToMemory(repoURL)
	if err == nil {
		return memFiles, nil
	}
	cloneErr := err
	log.Printf("Remote clone failed: %v\n", cloneErr)

	if isDefault {
		files, err := loadEmbeddedTemplate()
		if err != nil {
			log.Printf("Warning: embedded template invalid: %v\n", err)
		} else if files != nil {
			return files, nil
		}
	}

	return nil, fmt.Errorf("failed to load template: %w", cloneErr)
}

// 注意：getParentPath 定义在 db_handler.go 中
