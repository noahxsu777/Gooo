package events

import "time"

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
)

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Priority  Priority  `json:"priority"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "desconectado"
	StatusConnecting   ConnectionStatus = "conectando"
	StatusLive         ConnectionStatus = "live"
	StatusReconnecting ConnectionStatus = "reconectando"
)

type State struct {
	Status          ConnectionStatus `json:"status"`
	Channel         string           `json:"channel"`
	LatencyMS       int64            `json:"latencyMs"`
	QueueSize       int              `json:"queueSize"`
	QueueCapacity   int              `json:"queueCapacity"`
	MaxTextLength   int              `json:"maxTextLength"`
	LastEvent       *Message         `json:"lastEvent,omitempty"`
	LastError       string           `json:"lastError,omitempty"`
	Reconnects      int              `json:"reconnects"`
	DroppedMessages int64            `json:"droppedMessages"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}
