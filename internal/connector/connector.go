package connector

import (
	"context"

	"github.com/noahxsu777/Gooo/internal/events"
)

type Connector interface {
	Name() string
	Connect(ctx context.Context, channel string) (<-chan events.Message, <-chan error)
}

// Configurable endpoints/secrets for external connectors must be injected by env/UI.
type TikTokLiveAdapter struct {
	Endpoint string
	Token    string
}

type TikToolsAdapter struct {
	Endpoint string
	APIKey   string
}
