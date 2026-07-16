package webhook

import (
	"testing"
	"time"
)

func TestSignatureValidation(t *testing.T) {
	body := []byte(`{"a":1}`)
	sig := Sign("secret", body)
	if !Validate("secret", sig, body) {
		t.Fatal("expected valid signature")
	}
	if Validate("secret", "sha256=bad", body) {
		t.Fatal("expected invalid signature")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 10*time.Second)
	if !rl.Allow("k") || !rl.Allow("k") {
		t.Fatal("first two calls should pass")
	}
	if rl.Allow("k") {
		t.Fatal("third call should be limited")
	}
}
