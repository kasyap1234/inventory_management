package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyPaymentSignature(t *testing.T) {
	orderID := "order_123"
	paymentID := "pay_456"
	secret := "supersecret"

	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	svc := NewRazorpayService("key", secret, "hook")

	require.NoError(t, svc.VerifyPaymentSignature(orderID, paymentID, sig))
	require.Error(t, svc.VerifyPaymentSignature(orderID, paymentID, "bad-signature"))
}

func TestDetectContentType(t *testing.T) {
	require.Equal(t, "image/png", detectContentType(".png"))
	require.Equal(t, "image/jpeg", detectContentType(".jpeg"))
	require.Equal(t, "image/jpeg", detectContentType(".jpg"))
	require.Equal(t, "image/webp", detectContentType(".webp"))
	require.Equal(t, "image/gif", detectContentType(".gif"))
	require.Equal(t, "image/jpeg", detectContentType(".unknown"))
}
