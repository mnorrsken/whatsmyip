package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWhatsmyipSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WhatsMyIP Suite")
}
