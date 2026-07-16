package tts

import (
	"strings"

	"github.com/noahxsu777/Gooo/internal/events"
)

type Filters struct {
	MaxLength      int
	IgnoreCommands bool
	BlockedUsers   map[string]struct{}
	BlockedTypes   map[string]struct{}
	SpamWindow     int

	lastByUser map[string]string
}

func (f *Filters) ShouldSpeak(msg events.Message) bool {
	if len(strings.TrimSpace(msg.Text)) == 0 {
		return false
	}
	if f.MaxLength > 0 && len(msg.Text) > f.MaxLength {
		return false
	}
	if f.IgnoreCommands && strings.HasPrefix(strings.TrimSpace(msg.Text), "!") {
		return false
	}
	if _, blocked := f.BlockedUsers[strings.ToLower(msg.User)]; blocked {
		return false
	}
	if _, blocked := f.BlockedTypes[strings.ToLower(msg.Type)]; blocked {
		return false
	}
	if f.SpamWindow > 0 {
		if f.lastByUser == nil {
			f.lastByUser = map[string]string{}
		}
		u := strings.ToLower(msg.User)
		if f.lastByUser[u] == msg.Text {
			return false
		}
		f.lastByUser[u] = msg.Text
	}
	return true
}
