package tts

import (
	"testing"

	"github.com/noahxsu777/Gooo/internal/events"
)

func TestShouldSpeak(t *testing.T) {
	f := &Filters{
		MaxLength:      10,
		IgnoreCommands: true,
		BlockedUsers:   map[string]struct{}{"bad": {}},
		BlockedTypes:   map[string]struct{}{"gift": {}},
		SpamWindow:     1,
	}
	cases := []struct {
		name string
		msg  events.Message
		ok   bool
	}{
		{"ok", events.Message{User: "u", Text: "hola", Type: "chat"}, true},
		{"command", events.Message{User: "u", Text: "!ban", Type: "chat"}, false},
		{"blocked", events.Message{User: "bad", Text: "hola", Type: "chat"}, false},
		{"long", events.Message{User: "u", Text: "12345678901", Type: "chat"}, false},
		{"type", events.Message{User: "u", Text: "gift", Type: "gift"}, false},
	}
	for _, tc := range cases {
		if got := f.ShouldSpeak(tc.msg); got != tc.ok {
			t.Fatalf("%s expected %v got %v", tc.name, tc.ok, got)
		}
	}
	if f.ShouldSpeak(events.Message{User: "u", Text: "hola", Type: "chat"}) {
		t.Fatal("duplicate spam text should be blocked")
	}
}
