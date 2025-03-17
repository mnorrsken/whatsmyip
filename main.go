package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", handler)

	log.Printf("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s", r.RemoteAddr)
	log.Printf("Request method: %s, URL path: %s", r.Method, r.URL.Path)
	log.Printf("User agent: %s", r.UserAgent())

	clientIP := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	xForwardedHost := r.Header.Get("X-Forwarded-Host")
	xForwardedProto := r.Header.Get("X-Forwarded-Proto")
	host := r.Host

	log.Printf("Client IP: %s", clientIP)
	log.Printf("X-Forwarded-For: %s", xForwardedFor)
	log.Printf("X-Forwarded-Host: %s", xForwardedHost)
	log.Printf("X-Forwarded-Proto: %s", xForwardedProto)
	log.Printf("Host: %s", host)

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

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Request handled successfully")
}
