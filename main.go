// Licensed under GPL, 2016
// Refer to LICENSE for more details
// Refer to README for structural information

package main

import (
	"github.com/dchest/captcha"
	"html/template"
	"log"
	"net/http"
	"strk.kbt.io/projects/go/libravatar"
)

const (
	diff = "d"
	new  = "n"
	pop  = "p"
)

type errorPage struct {
	H, M string // header and message
}

var t *template.Template

func main() {
	t = template.Must(template.New("redb").Funcs(template.FuncMap{
		"libr": func(author string) string {
			r, _ := libravatar.FromEmail(author)
			return r
		},
	}).ParseGlob("html/*.gtml"))

	// close database connection before quitting
	defer db.Close()

	// start ReGeX game server in parallel
	go gameServer()

	// handle HTTP requests
	http.HandleFunc("/", index(new, true))
	http.HandleFunc("/new", index(new, false))
	http.HandleFunc("/pop", index(pop, false))
	http.HandleFunc("/diff", index(diff, false))
	http.HandleFunc("/r/", showRegex)
	http.HandleFunc("/contrib", contrib)
	http.HandleFunc("/search", search)
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "about.gtml", nil)
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./css/style.css")
	})
	http.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./img/logo.png")
	})
	http.HandleFunc("/flavicon.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./img/flavicon.png")
	})
	http.Handle("/c/", captcha.Server(260, 80))
	http.Handle("/source/", http.StripPrefix("/source/", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
