package main

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/jasonlabz/gentol/metadata"
	"os"
	"path/filepath"
)

var innerBox *packr.Box

// IsExist check file or directory
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsFile checks whether the path is a file,
// it returns false when it's a directory or does not exist.
func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

// IsDir checks whether the path is a directory,
// it returns false when it's a file or does not exist.
func IsDir(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return fi.IsDir()
}

// LoadTemplate return template from template dir, falling back to the embedded templates
func LoadTemplate(filename string, templateDir string) (tpl *metadata.Template, err error) {
	baseName := filepath.Base(filename)
	if templateDir != "" {
		fpath := filepath.Join("", filename)
		var b []byte
		b, err = os.ReadFile(fpath)
		if err == nil {
			absPath, err := filepath.Abs(fpath)
			if err != nil {
				absPath = fpath
			}
			tpl = &metadata.Template{Name: "file://" + absPath, Content: string(b)}
			return tpl, nil
		}
	}
	template, ok := metadata.LoadTpl(baseName)
	if !ok {
		return nil, fmt.Errorf("%s not found internally", baseName)
	}

	return template, nil
}

func init() {
	//innerBox = packr.New("gentol", "./template")
	//_, filename, _, ok := runtime.Caller(1)
	//if ok {
	//	fmt.Println(filename)
	//}
	//files, err := metadata.ListDir("./template", "")
	//if err != nil {
	//	panic(err)
	//}
	//for _, file := range files {
	//	baseName := filepath.Base(file)
	//	var b []byte
	//	b, err = os.ReadFile(file)
	//	if err == nil {
	//		metadata.StoreTpl(baseName, string(b))
	//	}
	//}
}
