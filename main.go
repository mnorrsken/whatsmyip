package main

import (
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	xForwardedHost := r.Header.Get("X-Forwarded-Host")
	xForwardedProto := r.Header.Get("X-Forwarded-Proto")
	host := r.Host

	data := struct {
		ClientIP        string
		XForwardedFor   string
		XForwardedHost  string
		XForwardedProto string
		Host            string
	}{
		ClientIP:        clientIP,
		XForwardedFor:   xForwardedFor,
		XForwardedHost:  xForwardedHost,
		XForwardedProto: xForwardedProto,
		Host:            host,
	}

	tmpl.Execute(w, data)
}
