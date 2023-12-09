package ui

import (
	"html/template"
	"net/http"
)

var index = template.Must(template.ParseFiles("ui/index.html", "ui/base.html"))
var loginForm = template.Must(template.ParseFiles("ui/login-form.html", "ui/base.html"))
var signupForm = template.Must(template.ParseFiles("ui/signup-form.html", "ui/base.html"))
var dashboard = template.Must(template.ParseFiles("ui/dashboard.html", "ui/base.html"))

var templates = map[string]*template.Template{
	"index":      index,
	"loginForm":  loginForm,
	"signupForm": signupForm,
	"dashboard":  dashboard,
}

func Render(w http.ResponseWriter, data interface{}, tmpl string) {
	t := templates[tmpl]
	err := t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
