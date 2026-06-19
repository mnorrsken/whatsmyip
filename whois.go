package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

const defaultWhoisBaseURL = "http://ip-api.com/json"

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
	cache   *ttlcache.Cache[string, *WhoisInfo]
	client  *http.Client
	baseURL string
}

// NewWhoisService initializes a WhoisService with sensible defaults.
func NewWhoisService(baseURL string) *WhoisService {
	normalized := strings.TrimSpace(baseURL)
	normalized = strings.TrimSuffix(normalized, "/")
	if normalized == "" {
		normalized = defaultWhoisBaseURL
	}

	c := ttlcache.New[string, *WhoisInfo](
		ttlcache.WithTTL[string, *WhoisInfo](1*time.Hour),
		ttlcache.WithDisableTouchOnHit[string, *WhoisInfo](),
	)
	go c.Start()

	return &WhoisService{
		cache:   c,
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: normalized,
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

	baseURL := s.baseURL
	if baseURL == "" {
		baseURL = defaultWhoisBaseURL
	}

	endpoint := fmt.Sprintf("%s/%s", baseURL, ipAddress)

	log.Printf("Fetching whois info for IP: %s", ipAddress)
	resp, err := client.Get(endpoint)
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

	if item := s.cache.Get(ipAddress); item != nil {
		log.Printf("Cache hit for IP: %s", ipAddress)
		return item.Value()
	}

	return nil
}

func (s *WhoisService) storeInCache(ipAddress string, info *WhoisInfo) {
	if s == nil || s.cache == nil || info == nil {
		return
	}

	s.cache.Set(ipAddress, info, ttlcache.DefaultTTL)
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
