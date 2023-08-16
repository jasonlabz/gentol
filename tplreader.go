package main

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"

	"os"
	"path/filepath"
	"sync"
)

var tplMap sync.Map
var innerBox *packr.Box

// Template template info struct
type Template struct {
	Name    string
	Content string
}

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
func LoadTemplate(filename string) (tpl *Template, err error) {
	baseName := filepath.Base(filename)
	if *templateDir != "" {
		fpath := filepath.Join("", filename)
		var b []byte
		b, err = os.ReadFile(fpath)
		if err == nil {
			absPath, err := filepath.Abs(fpath)
			if err != nil {
				absPath = fpath
			}
			tpl = &Template{Name: "file://" + absPath, Content: string(b)}
			return tpl, nil
		}
	}
	content, err := innerBox.FindString(baseName)
	if err != nil {
		return nil, fmt.Errorf("%s not found internally", baseName)
	}

	tpl = &Template{Name: "internal://" + filename, Content: content}
	return tpl, nil
}

func init() {
	innerBox = packr.New("gentol", "./template")
}
