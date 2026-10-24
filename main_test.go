package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testTemplate = template.Must(template.ParseFiles("templates/index.html"))

type stubWhoisService struct{}

func (stubWhoisService) Lookup(ip string) (*WhoisInfo, error) {
	if isPrivateIP(ip) {
		return &WhoisInfo{
			Status:     "success",
			Message:    "Private IP address",
			RegionName: "Local",
		}, nil
	}

	return &WhoisInfo{
		Status:      "success",
		Country:     "TestCountry",
		CountryCode: "TC",
		City:        "TestCity",
		ISP:         "TestISP",
		Org:         "TestOrg",
	}, nil
}

// proxyAddr is a private peer address, so the server trusts the
// X-Forwarded-For header it receives from it.
const proxyAddr = "10.0.0.2:41000"

func newTestServer() *Server {
	return NewServer(testTemplate, stubWhoisService{}, nil, nil)
}

var _ = Describe("HTTP Handler", func() {
	DescribeTable("Client IP and headers display",
		func(xForwardedFor, xForwardedHost, customHeader, expectedIPContains, expectedServer string) {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = proxyAddr

			// Set headers for test
			if xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", xForwardedFor)
			}
			if xForwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", xForwardedHost)
			}
			req.Header.Set("Custom-Test-Header", customHeader)
			// Add Remote-User header for all standard tests to pass with new requirement
			req.Header.Set("Remote-User", "testuser")

			rr := httptest.NewRecorder()
			server := newTestServer()

			server.ServeHTTP(rr, req)

			// Check status code
			Expect(rr.Code).To(Equal(http.StatusOK))

			responseBody := rr.Body.String()

			// Check for client IP
			if expectedIPContains != "<h1></h1>" {
				Expect(responseBody).To(ContainSubstring(expectedIPContains))
			}

			// Check for server information if expected
			if expectedServer != "" {
				Expect(responseBody).To(ContainSubstring(expectedServer))
			}

			// Check that custom header is included in response
			Expect(responseBody).To(ContainSubstring("Custom-Test-Header"))
			Expect(responseBody).To(ContainSubstring(customHeader))
		},
		Entry("No forwarded headers", "", "", "TestValue", "<h1", ""),
		Entry("With forwarded headers", "192.0.2.1", "example.com", "AnotherValue", "192.0.2.1", "example.com"),
	)

	Context("When handling specific header cases", func() {
		It("should include multiple headers in the response", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())

			// Set multiple test headers
			req.Header.Set("X-Test-1", "Value1")
			req.Header.Set("X-Test-2", "Value2")
			req.Header.Set("User-Agent", "GinkgoTest")
			req.Header.Set("Remote-User", "testuser")

			rr := httptest.NewRecorder()
			server := newTestServer()

			server.ServeHTTP(rr, req)

			responseBody := rr.Body.String()
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(responseBody).To(ContainSubstring("X-Test-1"))
			Expect(responseBody).To(ContainSubstring("Value1"))
			Expect(responseBody).To(ContainSubstring("X-Test-2"))
			Expect(responseBody).To(ContainSubstring("Value2"))
			Expect(responseBody).To(ContainSubstring("User-Agent"))
			Expect(responseBody).To(ContainSubstring("GinkgoTest"))
		})
	})

	Context("When handling Remote-User authentication", func() {
		It("should show only IP for non-private IP without Remote-User header", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = proxyAddr

			// Use a public IP address
			req.Header.Set("X-Forwarded-For", "8.8.8.8")
			// Intentionally not setting Remote-User

			rr := httptest.NewRecorder()
			server := newTestServer()

			server.ServeHTTP(rr, req)

			// Should still return 200 OK
			Expect(rr.Code).To(Equal(http.StatusOK))

			responseBody := rr.Body.String()
			// Should contain the IP address
			Expect(responseBody).To(ContainSubstring("8.8.8.8"))

			// Should not contain headers
			Expect(responseBody).NotTo(ContainSubstring("X-Forwarded-For"))

			// Should not contain whois information
			Expect(responseBody).NotTo(ContainSubstring("Country:"))
			Expect(responseBody).NotTo(ContainSubstring("ISP:"))
		})

		It("should allow access for non-private IP with Remote-User header", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = proxyAddr

			// Use a public IP address
			req.Header.Set("X-Forwarded-For", "8.8.8.8")
			req.Header.Set("Remote-User", "authenticateduser")

			rr := httptest.NewRecorder()
			server := newTestServer()

			server.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			responseBody := rr.Body.String()

			// Should contain the IP address
			Expect(responseBody).To(ContainSubstring("8.8.8.8"))

			// Should contain headers
			Expect(responseBody).To(ContainSubstring("X-Forwarded-For"))
			Expect(responseBody).To(ContainSubstring("Remote-User"))
		})

		It("should show full information for private IP without Remote-User header", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = proxyAddr

			// Use a private IP address
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			// Intentionally not setting Remote-User

			rr := httptest.NewRecorder()
			server := newTestServer()

			server.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			responseBody := rr.Body.String()

			// Should contain the IP address
			Expect(responseBody).To(ContainSubstring("192.168.1.1"))

			// Should contain headers
			Expect(responseBody).To(ContainSubstring("X-Forwarded-For"))

			// Should contain whois information
			Expect(responseBody).To(ContainSubstring("Private IP address"))
		})
	})

	Context("When X-Forwarded-For is forged", func() {
		It("should not show details to a public client claiming a private IP", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = proxyAddr
			// The proxy appends the real client address to what the client sent.
			req.Header.Set("X-Forwarded-For", "192.168.1.1, 203.0.113.9")

			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			responseBody := rr.Body.String()
			Expect(responseBody).To(ContainSubstring("203.0.113.9"))
			Expect(responseBody).To(ContainSubstring("Authentication required"))
			Expect(responseBody).NotTo(ContainSubstring("Private IP address"))
		})

		It("should ignore the header from a peer that is not a trusted proxy", func() {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = "203.0.113.9:52000"
			req.Header.Set("X-Forwarded-For", "10.0.0.1")

			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			responseBody := rr.Body.String()
			Expect(responseBody).To(ContainSubstring("203.0.113.9"))
			Expect(responseBody).To(ContainSubstring("Authentication required"))
		})
	})

	DescribeTable("Client IP resolution",
		func(remoteAddr string, forwardedFor []string, expected string) {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			req.RemoteAddr = remoteAddr
			for _, value := range forwardedFor {
				req.Header.Add("X-Forwarded-For", value)
			}

			Expect(newTestServer().resolveClientIP(req)).To(Equal(expected))
		},
		Entry("direct public peer", "203.0.113.9:52000", nil, "203.0.113.9"),
		Entry("direct public peer with forged header", "203.0.113.9:52000", []string{"10.0.0.1"}, "203.0.113.9"),
		Entry("proxy without header", proxyAddr, nil, "10.0.0.2"),
		Entry("proxy with public client", proxyAddr, []string{"203.0.113.9"}, "203.0.113.9"),
		Entry("proxy with private client", proxyAddr, []string{"192.168.1.20"}, "192.168.1.20"),
		Entry("forged entries left of the client", proxyAddr, []string{"10.0.0.1, 198.51.100.7, 203.0.113.9"}, "203.0.113.9"),
		Entry("chain of private proxies", proxyAddr, []string{"203.0.113.9, 10.42.0.5, 10.42.1.6"}, "203.0.113.9"),
		Entry("multiple header lines", proxyAddr, []string{"10.0.0.1", "203.0.113.9"}, "203.0.113.9"),
		Entry("IPv6 client", "[fd00::2]:41000", []string{"2001:db8::1"}, "2001:db8::1"),
		Entry("garbage right of the client", proxyAddr, []string{"203.0.113.9, not-an-ip"}, "10.0.0.2"),
		Entry("garbage left of the client", proxyAddr, []string{"not-an-ip, 203.0.113.9"}, "203.0.113.9"),
	)
})
