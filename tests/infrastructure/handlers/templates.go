package handlers

import (
	"html/template"
	"path/filepath"
	"runtime"
)

var Templates *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	tmplDir := filepath.Join(filepath.Dir(filename), "..", "templates", "*.tmpl")
	Templates = template.Must(template.ParseGlob(tmplDir))
}
