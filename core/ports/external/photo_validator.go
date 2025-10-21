package external

// PhotoValidationResult represents the result of validating a photo URL
type PhotoValidationResult struct {
	URL         string `json:"url"`
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
}

// PhotoValidator defines the interface for validating photo URLs with SSRF protection.
// Implements security requirements from FR-004:
// - Only HTTP and HTTPS protocols
// - No localhost, private IP ranges, or link-local addresses
// - 5 second timeout for accessibility checks
// - Only image content types (image/jpeg, image/png, image/webp)
type PhotoValidator interface {
	// ValidateURL checks if a single photo URL is valid, accessible, and secure.
	// Returns validation result with details about the check.
	ValidateURL(url string) PhotoValidationResult

	// ValidateURLs checks multiple photo URLs and returns results for each.
	// Validates 1-10 URLs per FR-004 requirement.
	ValidateURLs(urls []string) []PhotoValidationResult

	// IsSecureURL checks if URL passes SSRF protection without making HTTP requests.
	// Returns error if URL uses non-HTTP(S) protocol, points to private IPs, or localhost.
	IsSecureURL(url string) error
}
