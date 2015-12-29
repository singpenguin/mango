package mango

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/singpenguin/utils"
)

var Template map[string]*template.Template

func TemplateLoader(basepath string) map[string]*template.Template {
	Template = make(map[string]*template.Template)
	tmp := make(map[string]string)

	err := filepath.Walk(basepath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.HasPrefix(f.Name(), ".") {
			return nil
		}

		fileBytes, e := ioutil.ReadFile(path)
		if e != nil {
			fmt.Println("Failed reading file:", path)
			return nil
		}
		tmp[path[len(basepath)+1:]] = string(fileBytes)

		return nil
	})
	if err != nil {
		fmt.Println("filepath.Walk() returned ", err)
	}

	re, _ := regexp.Compile("\\{\\{[\\S\\s]+?\\}\\}")
	for k, _ := range tmp {
		m := re.FindAllString(tmp[k], -1)
		if len(m) > 0 {
			for j, _ := range m {
				s := strings.Trim(m[j], "{")
				s = strings.Trim(s, "}")
				s = strings.TrimSpace(s)
				if s == k {
					fmt.Println("Template Circular reference:", k)
					break
				}
				tmp[k] = strings.Replace(tmp[k], m[j], tmp[s], -1)
			}
		}
		Template[k], err = template.New(k).Parse(tmp[k])
		if err != nil {
			fmt.Println("Failed parse template:", k, err)
		}
	}
	return Template
}
