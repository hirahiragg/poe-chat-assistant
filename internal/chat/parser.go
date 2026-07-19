package chat

import (
	"regexp"
	"strings"
	"time"
)

// PoE Client.txt log line format:
//   2025/07/12 04:30:18 46850203 cff945b9 [INFO Client 51360] #<PTKFGS> Teairra_Merc: message
//   2025/07/12 04:46:18 47809953 cff945b9 [INFO Client 51360] @From <ANgRY> RuxMerc: message
//   2025/07/12 04:46:46 47838359 cff945b9 [INFO Client 51360] @To RuxMerc: message
//
// Chat prefixes:
//   #  = Global
//   $  = Trade
//   %  = Party
//   @From = Whisper (inbound)
//   @To   = Whisper (outbound)

var logLineRe = regexp.MustCompile(
	`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \d+ [0-9a-f]+ \[INFO Client \d+\] (.+)$`,
)

var chatPrefixes = []struct {
	prefix  string
	channel Channel
}{
	{"@From ", ChannelWhisperIn},
	{"@To ", ChannelWhisperOut},
	{"#", ChannelGlobal},
	{"$", ChannelTrade},
	{"%", ChannelParty},
}

var guildRe = regexp.MustCompile(`^<([^>]+)>\s*`)

func ParseLine(line string) (Message, bool) {
	m := logLineRe.FindStringSubmatch(line)
	if m == nil {
		return Message{}, false
	}

	ts, err := time.Parse("2006/01/02 15:04:05", m[1])
	if err != nil {
		return Message{}, false
	}

	payload := m[2]

	var channel Channel
	var rest string
	matched := false
	for _, cp := range chatPrefixes {
		if strings.HasPrefix(payload, cp.prefix) {
			channel = cp.channel
			rest = payload[len(cp.prefix):]
			matched = true
			break
		}
	}
	if !matched {
		return Message{}, false
	}

	guild := ""
	if gm := guildRe.FindStringSubmatch(rest); gm != nil {
		guild = gm[1]
		rest = rest[len(gm[0]):]
	}

	idx := strings.Index(rest, ": ")
	if idx < 0 {
		return Message{}, false
	}

	player := rest[:idx]
	body := rest[idx+2:]

	return Message{
		Timestamp: ts,
		Channel:   channel,
		Guild:     guild,
		Player:    player,
		Body:      body,
	}, true
}
