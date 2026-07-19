package translation

import (
	"context"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
)

type Direction string

const (
	Inbound  Direction = "inbound"
	Outbound Direction = "outbound"
)

type Request struct {
	Direction Direction
	Message   string
	Context   []chat.Message
}

type Translator interface {
	Translate(ctx context.Context, req Request) (string, error)
}
