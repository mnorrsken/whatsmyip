package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

// WhoisInfo stores IP whois lookup information.
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

// LookupService retrieves whois information for an IP address.
type LookupService interface {
	Lookup(ip string) (*WhoisInfo, error)
}

// WhoisService provides whois lookup backed by a cache and http client.
type WhoisService struct {
	cache  *cache.Cache
	client *http.Client
}

// NewWhoisService initializes a WhoisService with sensible defaults.
func NewWhoisService() *WhoisService {
	return &WhoisService{
		cache:  cache.New(1*time.Hour, 10*time.Minute),
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Lookup returns whois information for the provided IP address.
func (s *WhoisService) Lookup(ipAddress string) (*WhoisInfo, error) {
	if isPrivateIP(ipAddress) {
		return &WhoisInfo{
			Status:     "success",
			Message:    "Private IP address",
			RegionName: "Local",
		}, nil
	}

	if info := s.cacheLookup(ipAddress); info != nil {
		return info, nil
	}

	client := s.client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	log.Printf("Fetching whois info for IP: %s", ipAddress)
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

	s.storeInCache(ipAddress, &whoisInfo)
	log.Printf("Cached whois info for IP: %s", ipAddress)

	return &whoisInfo, nil
}

func (s *WhoisService) cacheLookup(ipAddress string) *WhoisInfo {
	if s == nil || s.cache == nil {
		return nil
	}

	if cachedInfo, found := s.cache.Get(ipAddress); found {
		if info, ok := cachedInfo.(*WhoisInfo); ok {
			log.Printf("Cache hit for IP: %s", ipAddress)
			return info
		}
	}

	return nil
}

func (s *WhoisService) storeInCache(ipAddress string, info *WhoisInfo) {
	if s == nil || s.cache == nil || info == nil {
		return
	}

	s.cache.Set(ipAddress, info, cache.DefaultExpiration)
}

// isPrivateIP checks if an IP address is in a well-known private range.
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		if ip4[0] == 127 {
			return true
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		if ip4[0] == 0 && ip4[1] == 0 && ip4[2] == 0 && ip4[3] == 0 {
			return true
		}
		return false
	}

	if ip.IsLoopback() {
		return true
	}
	if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
		return true
	}
	if (ip[0] & 0xfe) == 0xfc {
		return true
	}

	return false
}
