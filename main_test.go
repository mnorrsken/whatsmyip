package main

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Handler", func() {
	DescribeTable("Client IP and headers display",
		func(xForwardedFor, xForwardedHost, customHeader, expectedIPContains, expectedServer string) {
			req, err := http.NewRequest("GET", "/", nil)
			Expect(err).NotTo(HaveOccurred())

			// Set headers for test
			if xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", xForwardedFor)
			}
			if xForwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", xForwardedHost)
			}
			req.Header.Set("Custom-Test-Header", customHeader)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler)

			handler.ServeHTTP(rr, req)

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

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler)

			handler.ServeHTTP(rr, req)

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
})
