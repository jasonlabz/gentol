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
