package ui

import (
	"html/template"
	"net/http"
)

// func RenderTemplate(w http.ResponseWriter, tmpl string, data *interface{}) {
// 	t, err := template.ParseFiles("ui/" + tmpl + ".html")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	err = t.Execute(w, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

var login = template.Must(template.ParseFiles("ui/login.html", "ui/base.html"))
var dashboard = template.Must(template.ParseFiles("ui/dashboard.html", "ui/base.html"))

var templates = map[string]*template.Template{
	"login":     login,
	"dashboard": dashboard,
}

func RenderPage(w http.ResponseWriter, data interface{}, tmpl string) {
	t := templates[tmpl]
	err := t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
