package translation

import (
	"context"
	"testing"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
)

type mockTranslator struct {
	callCount int
	fn        func(req Request) string
}

func (m *mockTranslator) Translate(_ context.Context, req Request) (string, error) {
	m.callCount++
	return m.fn(req), nil
}

func TestServiceTranslate(t *testing.T) {
	mock := &mockTranslator{
		fn: func(req Request) string {
			if req.Direction == Inbound {
				return "翻訳結果: " + req.Message
			}
			return "translated: " + req.Message
		},
	}

	svc := NewService(mock)
	ctx := context.Background()

	t.Run("inbound translation", func(t *testing.T) {
		result, err := svc.Translate(ctx, Request{
			Direction: Inbound,
			Message:   "can you do 10 div?",
		})
		if err != nil {
			t.Fatal(err)
		}
		if result != "翻訳結果: can you do 10 div?" {
			t.Errorf("got %q", result)
		}
	})

	t.Run("outbound translation", func(t *testing.T) {
		result, err := svc.Translate(ctx, Request{
			Direction: Outbound,
			Message:   "大丈夫です",
		})
		if err != nil {
			t.Fatal(err)
		}
		if result != "translated: 大丈夫です" {
			t.Errorf("got %q", result)
		}
	})

	t.Run("cache hit does not call translator again", func(t *testing.T) {
		before := mock.callCount

		result, err := svc.Translate(ctx, Request{
			Direction: Inbound,
			Message:   "can you do 10 div?",
		})
		if err != nil {
			t.Fatal(err)
		}
		if result != "翻訳結果: can you do 10 div?" {
			t.Errorf("got %q", result)
		}
		if mock.callCount != before {
			t.Errorf("translator was called again (cache miss), callCount: %d -> %d", before, mock.callCount)
		}
	})

	t.Run("different direction is a cache miss", func(t *testing.T) {
		before := mock.callCount

		_, err := svc.Translate(ctx, Request{
			Direction: Outbound,
			Message:   "can you do 10 div?",
		})
		if err != nil {
			t.Fatal(err)
		}
		if mock.callCount != before+1 {
			t.Errorf("expected cache miss for different direction")
		}
	})

	t.Run("context changes cache key for outbound", func(t *testing.T) {
		before := mock.callCount

		_, err := svc.Translate(ctx, Request{
			Direction: Outbound,
			Message:   "大丈夫です",
			Context: []chat.Message{
				{Player: "PlayerA", Body: "how much?"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if mock.callCount != before+1 {
			t.Errorf("expected cache miss when context differs")
		}
	})
}
