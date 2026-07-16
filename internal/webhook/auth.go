package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Sign(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write(body)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func Validate(secret, signature string, body []byte) bool {
	if secret == "" || signature == "" {
		return false
	}
	expected := Sign(secret, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}
