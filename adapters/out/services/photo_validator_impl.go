package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// photoValidatorImpl implements external.PhotoValidator with SSRF protection
type photoValidatorImpl struct {
	httpClient *http.Client
}

// NewPhotoValidator creates a new PhotoValidator with 5-second timeout per FR-004
func NewPhotoValidator() external.PhotoValidator {
	return &photoValidatorImpl{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Prevent redirect loops (max 3 redirects)
				if len(via) >= 3 {
					return fmt.Errorf("stopped after 3 redirects")
				}
				// Validate redirect target for SSRF
				if err := validateURL(req.URL.String()); err != nil {
					return fmt.Errorf("unsafe redirect target: %w", err)
				}
				return nil
			},
		},
	}
}

// ValidateURL checks if a single photo URL is valid, accessible, and secure
func (v *photoValidatorImpl) ValidateURL(urlStr string) external.PhotoValidationResult {
	result := external.PhotoValidationResult{
		URL:   urlStr,
		Valid: false,
	}

	// Check SSRF protection
	if err := v.IsSecureURL(urlStr); err != nil {
		result.Error = err.Error()
		return result
	}

	// Make HEAD request to check accessibility and content type
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, urlStr, nil)
	if err != nil {
		result.Error = fmt.Sprintf("invalid URL: %v", err)
		return result
	}

	// Set user agent to identify our service
	req.Header.Set("User-Agent", "JalanRusak-PhotoValidator/1.0")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("URL not accessible: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Error = fmt.Sprintf("HTTP %d: URL not accessible", resp.StatusCode)
		return result
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !isValidImageContentType(contentType) {
		result.Error = fmt.Sprintf("invalid content type: %s (expected image/jpeg, image/png, or image/webp)", contentType)
		return result
	}

	// Get content length if available
	if contentLength := resp.ContentLength; contentLength > 0 {
		result.SizeBytes = contentLength
	}

	result.Valid = true
	result.ContentType = contentType
	return result
}

// ValidateURLs checks multiple photo URLs
func (v *photoValidatorImpl) ValidateURLs(urls []string) []external.PhotoValidationResult {
	results := make([]external.PhotoValidationResult, len(urls))
	for i, urlStr := range urls {
		results[i] = v.ValidateURL(urlStr)
	}
	return results
}

// IsSecureURL checks if URL passes SSRF protection
func (v *photoValidatorImpl) IsSecureURL(urlStr string) error {
	return validateURL(urlStr)
}

// validateURL performs comprehensive SSRF protection checks
func validateURL(urlStr string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check protocol (only HTTP and HTTPS allowed)
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid protocol: %s (only HTTP and HTTPS allowed)", parsed.Scheme)
	}

	// Extract hostname
	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("missing hostname")
	}

	// Block localhost and loopback
	if isLocalhost(hostname) {
		return fmt.Errorf("localhost and loopback addresses are not allowed (SSRF protection)")
	}

	// Resolve hostname to IP addresses
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if isPrivateOrReservedIP(ip) {
			return fmt.Errorf("private, reserved, or link-local IP addresses are not allowed: %s (SSRF protection)", ip.String())
		}
	}

	return nil
}

// isLocalhost checks if hostname is localhost or loopback
func isLocalhost(hostname string) bool {
	hostname = strings.ToLower(hostname)
	return hostname == "localhost" ||
		hostname == "127.0.0.1" ||
		hostname == "::1" ||
		strings.HasPrefix(hostname, "127.") ||
		hostname == "[::1]"
}

// isPrivateOrReservedIP checks if IP is private, link-local, or reserved
func isPrivateOrReservedIP(ip net.IP) bool {
	// Check for private IPv4 ranges (RFC 1918)
	privateIPv4Blocks := []string{
		"10.0.0.0/8",     // Private network
		"172.16.0.0/12",  // Private network
		"192.168.0.0/16", // Private network
		"169.254.0.0/16", // Link-local
		"127.0.0.0/8",    // Loopback
		"0.0.0.0/8",      // Current network
		"100.64.0.0/10",  // Shared address space
	}

	for _, cidr := range privateIPv4Blocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}

	// Check for private IPv6 ranges
	if ip.To4() == nil { // IPv6
		privateIPv6Blocks := []string{
			"::1/128",       // Loopback
			"fe80::/10",     // Link-local
			"fc00::/7",      // Unique local
			"ff00::/8",      // Multicast
			"::ffff:0:0/96", // IPv4-mapped IPv6
		}

		for _, cidr := range privateIPv6Blocks {
			_, block, _ := net.ParseCIDR(cidr)
			if block.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// isValidImageContentType checks if content type is an accepted image format
func isValidImageContentType(contentType string) bool {
	// Handle content types with charset or other parameters
	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	contentType = strings.TrimSpace(contentType)

	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}

	return false
}
