package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ComputeWebhookSignature computes HMAC-SHA256 over raw body bytes and returns "sha256=HEX_DIGEST"
// Returns empty string if secret is empty, but allows empty body for backwards compatibility
func ComputeWebhookSignature(body []byte, secret string) string {
	if strings.TrimSpace(secret) == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(secret))
	if len(body) > 0 {
		h.Write(body)
	}
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// VerifyWebhookSignature verifies the HMAC signature matches the expected value
func VerifyWebhookSignature(body []byte, secret string, providedSignature string) bool {
	if strings.TrimSpace(providedSignature) == "" {
		return false
	}
	expectedSignature := ComputeWebhookSignature(body, secret)
	return hmac.Equal([]byte(expectedSignature), []byte(providedSignature))
}

// ValidateOutgoingURLForWebhook enforces SSRF guard on target URL.
// - disallows loopback, link-local, and private ranges (IPv4 and IPv6)
// - restricts scheme to allowedSchemes (default ["https"])
// - if allowHTTPInDev is true and GO_ENV != production, http is allowed
// - prevents common SSRF attack vectors including URL manipulation
func ValidateOutgoingURLForWebhook(targetURL string, allowHTTPInDev bool, allowedSchemes []string) error {
	if strings.TrimSpace(targetURL) == "" {
		return fmt.Errorf("target_url is required")
	}

	// Check for maximum URL length to prevent DoS
	if len(targetURL) > 2048 {
		return fmt.Errorf("target_url exceeds maximum length of 2048 characters")
	}

	// Normalize and parse URL
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must include scheme and host")
	}

	// Prevent URL fragments and user info for security
	if u.Fragment != "" {
		return fmt.Errorf("URL fragments are not allowed in webhook URLs")
	}
	if u.User != nil {
		return fmt.Errorf("user credentials in URL are not allowed for security reasons")
	}

	// Scheme enforcement
	schemeAllowed := false
	for _, s := range allowedSchemes {
		if strings.EqualFold(u.Scheme, s) {
			schemeAllowed = true
			break
		}
	}
	// Allow HTTP in dev if env permits and not production
	goEnv := os.Getenv("GO_ENV")
	isProd := goEnv == "production" || goEnv == "prod"
	if !schemeAllowed {
		if strings.EqualFold(u.Scheme, "http") && allowHTTPInDev && !isProd {
			schemeAllowed = true
		}
	}
	if !schemeAllowed {
		return fmt.Errorf("scheme '%s' not allowed", u.Scheme)
	}

	// Use URL Hostname to properly extract host (handles IPv6 brackets and ports)
	host := u.Hostname()

	// Validate host is not empty after extraction
	if host == "" {
		return fmt.Errorf("invalid or missing hostname")
	}

	// Check for port if present - reject suspicious ports
	if port := u.Port(); port != "" {
		// Common unsafe ports to block
		unsafePorts := map[string]bool{
			"22": true, "23": true, "25": true, "110": true, "143": true, // SSH, Telnet, SMTP, POP3, IMAP
			"3306": true, "5432": true, "6379": true, "27017": true, // MySQL, PostgreSQL, Redis, MongoDB
		}
		if unsafePorts[port] {
			return fmt.Errorf("port %s is not allowed for webhook URLs", port)
		}
	}

	// If host is a literal IP, validate directly without DNS
	if ip := net.ParseIP(host); ip != nil {
		if isDeniedIP(ip) {
			goEnv := os.Getenv("GO_ENV")
			isProd := goEnv == "production" || goEnv == "prod"
			if allowHTTPInDev && !isProd && ip.IsLoopback() {
				// allow loopback in dev when explicitly enabled
			} else {
				return fmt.Errorf("ssrf_blocked: host resolves to a denied range (%s)", ip.String())
			}
		}
		return nil
	}

	// DNS/host resolution with timeout to prevent hanging
	ips, err := net.LookupIP(host)
	if err != nil {
		// Map DNS errors distinctly for caller
		return fmt.Errorf("dns_resolution_failed: %w", err)
	}

	// Ensure at least one IP was resolved
	if len(ips) == 0 {
		return fmt.Errorf("dns_resolution_failed: no IP addresses found for host")
	}

	// Deny ranges (optionally allow loopback in dev when explicitly enabled)
	goEnv = os.Getenv("GO_ENV")
	isProd = goEnv == "production" || goEnv == "prod"

	for _, ip := range ips {
		if isDeniedIP(ip) {
			// In non-production and when explicitly allowed for dev, permit loopback only
			if allowHTTPInDev && !isProd && ip.IsLoopback() {
				continue
			}
			return fmt.Errorf("ssrf_blocked: host resolves to a denied range (%s)", ip.String())
		}
	}

	return nil
}

func isDeniedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// Loopback
	if ip.IsLoopback() {
		return true
	}
	// Link-local unicast (169.254.0.0/16) or IPv6 fe80::/10
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// Multicast addresses
	if ip.IsMulticast() {
		return true
	}
	// Unspecified address (0.0.0.0 or ::)
	if ip.IsUnspecified() {
		return true
	}

	// Private ranges IPv4: 10/8, 172.16/12, 192.168/16
	if ip.To4() != nil {
		v4 := ip.To4()
		if v4[0] == 10 { // 10.0.0.0/8
			return true
		}
		if v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31 { // 172.16.0.0/12
			return true
		}
		if v4[0] == 192 && v4[1] == 168 { // 192.168.0.0/16
			return true
		}
		if v4[0] == 169 && v4[1] == 254 { // 169.254.0.0/16 link-local
			return true
		}
		if v4[0] == 127 { // 127.0.0.0/8 loopback (additional check)
			return true
		}
		// Broadcast address 255.255.255.255
		if v4[0] == 255 && v4[1] == 255 && v4[2] == 255 && v4[3] == 255 {
			return true
		}
		// RFC 1918 Class E (240.0.0.0/4) - experimental/reserved
		if v4[0] >= 240 {
			return true
		}
		return false
	}

	// IPv6 ::1 loopback
	if ip.Equal(net.ParseIP("::1")) {
		return true
	}
	// IPv6 ULA fc00::/7
	// Check first 7 bits = 0b1111110 -> fc00::/7 means first hex nibble 0xFC or 0xFD
	b := ip.To16()
	if b != nil {
		if b[0] == 0xfc || b[0] == 0xfd {
			return true
		}
		// IPv6 link-local fe80::/10
		if b[0] == 0xfe && (b[1]&0xc0) == 0x80 {
			return true
		}
	}
	return false
}

// ExtractWhitelistedHeaders returns a map of response headers whitelisted for UI
// Whitelist: content-type, date, server, x-request-id, content-length, via
func ExtractWhitelistedHeaders(resp *http.Response) map[string]string {
	whitelist := map[string]struct{}{
		"content-type":   {},
		"date":           {},
		"server":         {},
		"x-request-id":   {},
		"content-length": {},
		"via":            {},
	}
	out := make(map[string]string)
	for k, vals := range resp.Header {
		lk := strings.ToLower(k)
		if _, ok := whitelist[lk]; ok {
			if len(vals) > 0 {
				out[lk] = vals[0]
			}
		}
	}
	return out
}

// TruncateBodySnippet truncates a byte slice to max and returns string
func TruncateBodySnippet(b []byte, max int) string {
	if max <= 0 {
		return ""
	}
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max])
}

// BuildNoRedirectClient returns an http.Client that disallows redirects by default,
// or allows up to maxRedirects with per-redirect SSRF re-validation via validateFunc.
func BuildNoRedirectClient(timeout time.Duration, maxRedirects int, validateFunc func(location *url.URL) error) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if maxRedirects <= 0 {
				return errors.New("redirects_not_allowed")
			}
			if len(via) > maxRedirects {
				return errors.New("max_redirects_exceeded")
			}
			// Re-validate each redirect location
			if validateFunc != nil {
				loc := req.URL
				if err := validateFunc(loc); err != nil {
					return fmt.Errorf("redirect_ssrf_blocked: %w", err)
				}
			}
			return nil
		},
	}
}
