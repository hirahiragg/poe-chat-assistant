package chat

import (
	"fmt"
	"time"
)

type Channel int

const (
	ChannelGlobal     Channel = iota // #
	ChannelTrade                     // $
	ChannelParty                     // %
	ChannelGuild                     // &
	ChannelWhisperIn                 // @From
	ChannelWhisperOut                // @To
)

func (c Channel) String() string {
	switch c {
	case ChannelGlobal:
		return "Global"
	case ChannelTrade:
		return "Trade"
	case ChannelParty:
		return "Party"
	case ChannelGuild:
		return "Guild"
	case ChannelWhisperIn:
		return "Whisper"
	case ChannelWhisperOut:
		return "Whisper(out)"
	default:
		return "Unknown"
	}
}

func (c Channel) Symbol() string {
	switch c {
	case ChannelGlobal:
		return "#"
	case ChannelTrade:
		return "$"
	case ChannelParty:
		return "%"
	case ChannelGuild:
		return "&"
	case ChannelWhisperIn:
		return "@"
	case ChannelWhisperOut:
		return "→"
	default:
		return "?"
	}
}

type Message struct {
	Timestamp time.Time
	Channel   Channel
	Guild     string
	Player    string
	Body      string
}

func (m Message) String() string {
	guild := ""
	if m.Guild != "" {
		guild = fmt.Sprintf("<%s> ", m.Guild)
	}
	return fmt.Sprintf("[%s] %s %s%s: %s",
		m.Timestamp.Format("15:04:05"),
		m.Channel.Symbol(),
		guild,
		m.Player,
		m.Body,
	)
}
