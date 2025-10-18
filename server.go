package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Server handles incoming HTTP requests and renders template output.
type Server struct {
	tmpl    *template.Template
	whois   LookupService
	include map[string]struct{}
	exclude map[string]struct{}
}

// headerEntry captures a header name/value pair for template rendering.
type headerEntry struct {
	Name  string
	Value string
}

type templateData struct {
	ClientIP      string
	Headers       []headerEntry
	WhoisInfo     *WhoisInfo
	Authenticated bool
}

// NewServer creates a configured Server instance.
func NewServer(tmpl *template.Template, whois LookupService, includeHeaders, excludeHeaders map[string]bool) *Server {
	return &Server{
		tmpl:    tmpl,
		whois:   whois,
		include: toSet(includeHeaders),
		exclude: toSet(excludeHeaders),
	}
}

func toSet(values map[string]bool) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(values))
	for raw, enabled := range values {
		if !enabled {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" {
			continue
		}
		set[key] = struct{}{}
	}

	if len(set) == 0 {
		return nil
	}

	return set
}

// ServeHTTP implements http.Handler for the Server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.tmpl == nil {
		http.Error(w, "Template not configured", http.StatusInternalServerError)
		return
	}

	timeFormatted := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	requestLine := fmt.Sprintf("%s %s %s", r.Method, r.URL.RequestURI(), r.Proto)
	statusCode := 200
	contentLength := "-"

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "-"
	}

	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}

	clientIP := s.resolveClientIP(r)
	isPrivate := isPrivateIP(clientIP)
	isAuthenticated := isPrivate || r.Header.Get("Remote-User") != ""

	username := r.Header.Get("Remote-User")
	if username == "" {
		username = "-"
	}

	log.Printf("%s - %s [%s] \"%s\" %d %s \"%s\" \"%s\"",
		clientIP, username, timeFormatted, requestLine,
		statusCode, contentLength, referer, userAgent)

	var (
		whoisInfo *WhoisInfo
		headers   []headerEntry
	)

	if isAuthenticated {
		var err error
		whoisInfo, err = s.lookupWhois(clientIP)
		if err != nil {
			log.Printf("Error getting whois info: %v", err)
		}
		headers = s.collectHeaders(r.Header)
	} else {
		log.Printf("Unauthenticated access from non-private IP %s: Missing Remote-User header", clientIP)
	}

	data := templateData{
		ClientIP:      clientIP,
		Headers:       headers,
		WhoisInfo:     whoisInfo,
		Authenticated: isAuthenticated,
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Request handled successfully")
}

func (s *Server) lookupWhois(ip string) (*WhoisInfo, error) {
	if s == nil || s.whois == nil {
		return nil, nil
	}
	return s.whois.Lookup(ip)
}

func (s *Server) collectHeaders(header http.Header) []headerEntry {
	var entries []headerEntry
	for name, values := range header {
		lower := strings.ToLower(name)
		if len(s.include) > 0 {
			if _, ok := s.include[lower]; !ok {
				continue
			}
		}
		if _, ok := s.exclude[lower]; ok {
			continue
		}
		for _, value := range values {
			entries = append(entries, headerEntry{
				Name:  name,
				Value: value,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

func (s *Server) resolveClientIP(r *http.Request) string {
	clientIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}
	return clientIP
}
