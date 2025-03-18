package main

import (
	"html/template"
	"log"
	"net/http"
	"sort"
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

	// Get client IP from X-Forwarded-For header or fall back to RemoteAddr
	clientIP := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	// Get current server from X-Forwarded-Host header
	currentServer := r.Header.Get("X-Forwarded-Host")

	// Collect all headers
	var headers []struct {
		Name  string
		Value string
	}

	for name, values := range r.Header {
		for _, value := range values {
			headers = append(headers, struct {
				Name  string
				Value string
			}{
				Name:  name,
				Value: value,
			})
		}
	}

	// Sort headers alphabetically by name
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Name < headers[j].Name
	})

	data := struct {
		ClientIP      string
		CurrentServer string
		Headers       []struct {
			Name  string
			Value string
		}
	}{
		ClientIP:      clientIP,
		CurrentServer: currentServer,
		Headers:       headers,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Request handled successfully")
}
