package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

// RenderingTemplate rendering a template with data
func RenderingTemplate(templateInfo *Template, data map[string]any, outFilePath string, overwrite bool) (err error) {
	var file *os.File
	if !IsExist(outFilePath) && !overwrite {
		file, err = os.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("open file error %s\n", err.Error())
			return
		}
	}
	if overwrite {
		file, err = os.OpenFile(outFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("overwrite true: open file error %s\n", err.Error())
			return
		}
	}
	fileName := filepath.Base(outFilePath)

	tmpl, err := template.New(fileName).Option("missingkey=error").Parse(templateInfo.Content)
	if err != nil {
		return
	}
	//if err != nil {
	//	return fmt.Errorf("error in loading %s template, error: %v", genTemplate.Name, err)
	//}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error in rendering %s: %s", templateInfo.Name, err.Error())
	}

	fileContents, err := Format(templateInfo, buf.Bytes(), outFilePath)
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outFilePath, err)
	}

	_, err = io.Copy(file, bytes.NewReader(fileContents))
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outFilePath, err)
	}

	fmt.Printf("writing %s\n", outFilePath)

	return nil
}

func Format(templateInfo *Template, content []byte, outputFile string) ([]byte, error) {
	extension := filepath.Ext(outputFile)
	if extension == ".go" {
		formattedSource, err := format.Source([]byte(content))
		if err != nil {
			return nil, fmt.Errorf("error in formatting template: %s outputfile: %s source: %s", templateInfo.Name, outputFile, err.Error())
		}

		fileContents := NormalizeNewlines(formattedSource)
		fileContents = CRLFNewlines(formattedSource)
		return fileContents, nil
	}

	fileContents := NormalizeNewlines([]byte(content))
	fileContents = CRLFNewlines(fileContents)
	return fileContents, nil
}

// NormalizeNewlines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewlines(d []byte) []byte {
	// replace CR LF \r\n (windows) with LF \n (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (mac) with LF \n (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}

// CRLFNewlines transforms \n to \r\n (windows)
func CRLFNewlines(d []byte) []byte {
	// replace LF (unix) with CR LF \r\n (windows)
	d = bytes.Replace(d, []byte{10}, []byte{13, 10}, -1)
	return d
}
