package ui

import (
	"html/template"
	"net/http"
)

var landing = template.Must(template.ParseFiles("ui/pages/landing.html", "ui/base.html"))
var dashboard = template.Must(template.ParseFiles("ui/pages/dashboard.html", "ui/base.html"))
var chatroom = template.Must(template.ParseFiles("ui/pages/chatroom.html", "ui/base.html"))

var ptemplates = map[string]*template.Template{
	"landing":   landing,
	"dashboard": dashboard,
	"chatroom":  chatroom,
}

func RenderPage(w http.ResponseWriter, data interface{}, tmpl string) {
	t := ptemplates[tmpl]
	err := t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var ctemplates = template.Must(template.ParseFiles(
	"ui/components/login-form.html",
	"ui/components/signup-form.html",
))

func RenderComponent(w http.ResponseWriter, data interface{}, tmpl string) {
	err := ctemplates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
