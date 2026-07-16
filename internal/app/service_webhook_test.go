package app

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/noahxsu777/Gooo/internal/config"
	"github.com/noahxsu777/Gooo/internal/webhook"
)

func newTestService() *Service {
	cfg := config.Config{WebhookSecret: "secret", TikToolsWebhookHeader: "X-Tiktools-Signature", QueueSize: 5, MaxPayloadBytes: 128, Mode: "demo"}
	return NewService(cfg, slog.Default())
}

func TestWebhookAuthAndValidation(t *testing.T) {
	s := newTestService()
	body := []byte(`{"id":"1","type":"chat","user":"alice","text":"hola","priority":1}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/tiktools", bytes.NewReader(body))
	req.Header.Set("X-Tiktools-Signature", webhook.Sign("secret", body))
	res := httptest.NewRecorder()
	s.HandleWebhook(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got %d", res.Code)
	}

	reqBad := httptest.NewRequest(http.MethodPost, "/webhook/tiktools", bytes.NewReader(body))
	reqBad.Header.Set("X-Tiktools-Signature", "sha256=bad")
	resBad := httptest.NewRecorder()
	s.HandleWebhook(resBad, reqBad)
	if resBad.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", resBad.Code)
	}
}

func TestWebhookMaxPayload(t *testing.T) {
	s := newTestService()
	body := bytes.Repeat([]byte("a"), 256)
	req := httptest.NewRequest(http.MethodPost, "/webhook/tiktools", bytes.NewReader(body))
	req.Header.Set("X-Tiktools-Signature", webhook.Sign("secret", body))
	res := httptest.NewRecorder()
	s.HandleWebhook(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}
