package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name               string
		xForwardedFor      string
		xForwardedHost     string
		customHeader       string
		expectedIPContains string
		expectedServer     string
	}{
		{
			name:               "No forwarded headers",
			xForwardedFor:      "",
			xForwardedHost:     "",
			customHeader:       "TestValue",
			expectedIPContains: "<h1></h1>", // httptest uses local address
			expectedServer:     "",
		},
		{
			name:               "With forwarded headers",
			xForwardedFor:      "192.0.2.1",
			xForwardedHost:     "example.com",
			customHeader:       "AnotherValue",
			expectedIPContains: "192.0.2.1",
			expectedServer:     "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set headers for test
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xForwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tt.xForwardedHost)
			}
			req.Header.Set("Custom-Test-Header", tt.customHeader)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			responseBody := rr.Body.String()

			// Check for client IP
			if !strings.Contains(responseBody, tt.expectedIPContains) {
				t.Errorf("handler response doesn't contain expected IP %q\nGot response body:\n%s",
					tt.expectedIPContains, responseBody)
			}

			// Check for server information if expected
			if tt.expectedServer != "" && !strings.Contains(responseBody, tt.expectedServer) {
				t.Errorf("handler response doesn't contain expected server %q\nGot response body:\n%s",
					tt.expectedServer, responseBody)
			}

			// Check that custom header is included in response
			if !strings.Contains(responseBody, "Custom-Test-Header") ||
				!strings.Contains(responseBody, tt.customHeader) {
				t.Errorf("handler response doesn't contain expected custom header\nGot response body:\n%s",
					responseBody)
			}
		})
	}
}
