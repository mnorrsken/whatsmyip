package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "WhatsMyIP - A simple service to display client IP and HTTP headers\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	help := flag.Bool("help", false, "Show help message")
	port := flag.String("port", "8080", "Port to listen on")
	host := flag.String("host", "0.0.0.0", "IP address to listen on")
	includeFlag := flag.String("include", "", "Comma separated header names to include")
	excludeFlag := flag.String("exclude", "", "Comma separated header names to exclude")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	portNum, err := strconv.Atoi(*port)
	if err != nil || portNum <= 0 || portNum > 65535 {
		log.Fatalf("Invalid port: %s", *port)
	}

	includeHeaders := parseHeaderList(*includeFlag)
	excludeHeaders := parseHeaderList(*excludeFlag)

	whoisBaseURL := os.Getenv("WHOIS_BASE_URL")

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	server := NewServer(tmpl, NewWhoisService(whoisBaseURL), includeHeaders, excludeHeaders)

	listenAddr := *host + ":" + *port

	log.Printf("Starting server on http://%s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, server))
}

func parseHeaderList(raw string) map[string]bool {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	headers := make(map[string]bool)
	for _, item := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSpace(item))
		if name != "" {
			headers[name] = true
		}
	}

	if len(headers) == 0 {
		return nil
	}

	return headers
}
