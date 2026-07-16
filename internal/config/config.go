package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr                  string
	Channel               string
	Mode                  string
	WebhookSecret         string
	AllowedOrigin         string
	MaxPayloadBytes       int64
	QueueSize             int
	ReconnectBase         time.Duration
	ReconnectMax          time.Duration
	ReconnectMaxAttempts  int
	HeartbeatTimeout      time.Duration
	DemoDropEvery         int
	DemoInterval          time.Duration
	TikTokEndpoint        string
	TikToolsEndpoint      string
	TikToolsWebhookHeader string
}

func Load() Config {
	cfg := Config{
		Addr:                  getEnv("ADDR", ":8080"),
		Channel:               getEnv("TIKTOK_CHANNEL", "demo_canal"),
		Mode:                  getEnv("APP_MODE", "demo"),
		WebhookSecret:         getEnv("WEBHOOK_SECRET", "change-me-demo-secret"),
		AllowedOrigin:         getEnv("ALLOWED_ORIGIN", "*"),
		MaxPayloadBytes:       getInt64("MAX_PAYLOAD_BYTES", 65536),
		QueueSize:             getInt("QUEUE_SIZE", 120),
		ReconnectBase:         getDuration("RECONNECT_BASE", 1*time.Second),
		ReconnectMax:          getDuration("RECONNECT_MAX", 20*time.Second),
		ReconnectMaxAttempts:  getInt("RECONNECT_MAX_ATTEMPTS", 0),
		HeartbeatTimeout:      getDuration("HEARTBEAT_TIMEOUT", 20*time.Second),
		DemoDropEvery:         getInt("DEMO_DROP_EVERY", 18),
		DemoInterval:          getDuration("DEMO_INTERVAL", 1200*time.Millisecond),
		TikTokEndpoint:        getEnv("TIKTOK_CONNECTOR_ENDPOINT", ""),
		TikToolsEndpoint:      getEnv("TIKTOOLS_ENDPOINT", ""),
		TikToolsWebhookHeader: getEnv("WEBHOOK_SIGNATURE_HEADER", "X-Tiktools-Signature"),
	}
	if cfg.QueueSize < 1 {
		cfg.QueueSize = 1
	}
	return cfg
}

func getEnv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func getInt(k string, d int) int {
	if v, err := strconv.Atoi(getEnv(k, "")); err == nil {
		return v
	}
	return d
}

func getInt64(k string, d int64) int64 {
	if v, err := strconv.ParseInt(getEnv(k, ""), 10, 64); err == nil {
		return v
	}
	return d
}

func getDuration(k string, d time.Duration) time.Duration {
	if v, err := time.ParseDuration(getEnv(k, "")); err == nil {
		return v
	}
	return d
}
