package common

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeWebhookSignature_FormatAndDeterminism(t *testing.T) {
	body := []byte(`{"a":1,"b":"x"}`)
	secret := "topsecret"
	s1 := ComputeWebhookSignature(body, secret)
	s2 := ComputeWebhookSignature(body, secret)
	assert.Equal(t, s1, s2)
	assert.True(t, strings.HasPrefix(s1, "sha256="))
	assert.Len(t, s1, len("sha256=")+64) // hex-encoded 32-byte digest
}

func TestValidateOutgoingURLForWebhook_DisallowPrivateAndLoopback(t *testing.T) {
	os.Setenv("GO_ENV", "development") // ensure not prod for tests that enable http in dev

	// Loopback IPv4 - use allowed scheme to test SSRF blocking specifically
	err := ValidateOutgoingURLForWebhook("https://127.0.0.1:8080/hook", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// Localhost hostname
	err = ValidateOutgoingURLForWebhook("https://localhost/hook", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// Link-local - use allowed scheme to test SSRF blocking specifically
	err = ValidateOutgoingURLForWebhook("https://169.254.10.5/hook", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// Private 10/8
	err = ValidateOutgoingURLForWebhook("https://10.1.2.3/path", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// Private 172.16/12
	err = ValidateOutgoingURLForWebhook("https://172.16.5.5/path", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// Private 192.168/16
	err = ValidateOutgoingURLForWebhook("https://192.168.1.10/path", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// IPv6 loopback - use allowed scheme to test SSRF blocking specifically
	err = ValidateOutgoingURLForWebhook("https://[::1]/x", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")

	// IPv6 ULA fc00::/7 - use allowed scheme to test SSRF blocking specifically
	err = ValidateOutgoingURLForWebhook("https://[fd00::1]/x", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf_blocked")
}

func TestValidateOutgoingURLForWebhook_SchemeEnforcement(t *testing.T) {
	// HTTPS allowed by default
	err := ValidateOutgoingURLForWebhook("https://example.com/hook", false, []string{"https"})
	assert.NoError(t, err)

	// HTTP disallowed when not permitted
	err = ValidateOutgoingURLForWebhook("http://example.com/hook", false, []string{"https"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")

	// HTTP allowed in dev when flag enabled
	os.Setenv("GO_ENV", "development")
	err = ValidateOutgoingURLForWebhook("http://example.com/hook", true, []string{"https"})
	assert.NoError(t, err)

	// HTTPS-only in production is enforced by config, but util also respects GO_ENV for allowHTTPInDev
	os.Setenv("GO_ENV", "production")
	err = ValidateOutgoingURLForWebhook("http://example.com/hook", true, []string{"https"})
	assert.Error(t, err)
}
