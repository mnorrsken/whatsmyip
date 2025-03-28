package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WhoisInfo stores IP whois lookup information
type WhoisInfo struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// Global maps for header filtering.
var includeHeadersMap = make(map[string]bool)
var excludeHeadersMap = make(map[string]bool)

func main() {
	// Define custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "WhatsMyIP - A simple service to display client IP and HTTP headers\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Define flags
	help := flag.Bool("help", false, "Show help message")
	port := flag.String("port", "8080", "Port to listen on")
	host := flag.String("host", "0.0.0.0", "IP address to listen on")
	includeFlag := flag.String("include", "", "Comma separated header names to include")
	excludeFlag := flag.String("exclude", "", "Comma separated header names to exclude")
	flag.Parse()

	// Show help and exit if requested
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// validate port flag
	portNum, err := strconv.Atoi(*port)
	if err != nil || portNum <= 0 || portNum > 65535 {
		log.Fatalf("Invalid port: %s", *port)
	}

	// Process include flag
	if *includeFlag != "" {
		for _, h := range strings.Split(*includeFlag, ",") {
			includeHeadersMap[strings.ToLower(strings.TrimSpace(h))] = true
		}
	}

	// Process exclude flag
	if *excludeFlag != "" {
		for _, h := range strings.Split(*excludeFlag, ",") {
			excludeHeadersMap[strings.ToLower(strings.TrimSpace(h))] = true
		}
	}

	http.HandleFunc("/", handler)

	listenAddr := *host + ":" + *port
	log.Printf("Starting server on http://%s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func getWhoisInfo(ipAddress string) (*WhoisInfo, error) {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Call ip-api.com for whois information
	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s", ipAddress))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var whoisInfo WhoisInfo
	if err := json.Unmarshal(body, &whoisInfo); err != nil {
		return nil, err
	}

	return &whoisInfo, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s", r.RemoteAddr)
	log.Printf("Request method: %s, URL path: %s", r.Method, r.URL.Path)
	log.Printf("User agent: %s", r.UserAgent())

	// Get client IP from X-Forwarded-For header or fall back to RemoteAddr
	clientIP := r.RemoteAddr
	// Remove port from RemoteAddr if present
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	// Get whois information for the IP
	whoisInfo, err := getWhoisInfo(clientIP)
	if err != nil {
		log.Printf("Error getting whois info: %v", err)
		// Continue without whois info if there's an error
		whoisInfo = nil
	}

	// Collect all headers with filtering logic
	var headers []struct {
		Name  string
		Value string
	}
	for name, values := range r.Header {
		lname := strings.ToLower(name)
		if len(includeHeadersMap) > 0 && !includeHeadersMap[lname] {
			continue // skip if not in include list
		}
		if len(excludeHeadersMap) > 0 && excludeHeadersMap[lname] {
			continue // skip if in exclude list
		}
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
		ClientIP string
		Headers  []struct {
			Name  string
			Value string
		}
		WhoisInfo *WhoisInfo
	}{
		ClientIP:  clientIP,
		Headers:   headers,
		WhoisInfo: whoisInfo,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Request handled successfully")
}
