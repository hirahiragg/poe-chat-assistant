package chat

import (
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Message
		wantOK  bool
	}{
		{
			name:  "global chat",
			input: `2025/07/12 04:26:51 46643250 cff945b9 [INFO Client 51360] #ReallyFurious: sec`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 4, 26, 51, 0, time.UTC),
				Channel:   ChannelGlobal,
				Player:    "ReallyFurious",
				Body:      "sec",
			},
			wantOK: true,
		},
		{
			name:  "global chat with guild",
			input: `2025/07/12 04:30:18 46850203 cff945b9 [INFO Client 51360] #<PTKFGS> Teairra_Merc: i swear, maps influence what mercs u find`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 4, 30, 18, 0, time.UTC),
				Channel:   ChannelGlobal,
				Guild:     "PTKFGS",
				Player:    "Teairra_Merc",
				Body:      "i swear, maps influence what mercs u find",
			},
			wantOK: true,
		},
		{
			name:  "whisper inbound with guild",
			input: `2025/07/12 04:46:18 47809953 cff945b9 [INFO Client 51360] @From <ANgRY> RuxMerc: Hi, I would like to buy your level 3 20% Enlighten Support listed for 4 divine in Mercenaries (stash tab "Trade"; position: left 11, top 12)`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 4, 46, 18, 0, time.UTC),
				Channel:   ChannelWhisperIn,
				Guild:     "ANgRY",
				Player:    "RuxMerc",
				Body:      `Hi, I would like to buy your level 3 20% Enlighten Support listed for 4 divine in Mercenaries (stash tab "Trade"; position: left 11, top 12)`,
			},
			wantOK: true,
		},
		{
			name:  "whisper inbound without guild",
			input: `2025/07/12 05:44:37 51308796 cff945b9 [INFO Client 51360] @From DarkHolyhole: Hi, I would like to buy your level 3 4% Enlighten Support listed for 100 chaos`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 5, 44, 37, 0, time.UTC),
				Channel:   ChannelWhisperIn,
				Player:    "DarkHolyhole",
				Body:      "Hi, I would like to buy your level 3 4% Enlighten Support listed for 100 chaos",
			},
			wantOK: true,
		},
		{
			name:  "whisper outbound",
			input: `2025/07/12 04:46:46 47838359 cff945b9 [INFO Client 51360] @To RuxMerc: ty`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 4, 46, 46, 0, time.UTC),
				Channel:   ChannelWhisperOut,
				Player:    "RuxMerc",
				Body:      "ty",
			},
			wantOK: true,
		},
		{
			name:  "trade chat",
			input: `2025/07/14 17:52:51 267802812 cff945b9 [INFO Client 119368] $SumoUndies: any1 could help me get trough campaign starting at act 6 ? i can pay :)`,
			want: Message{
				Timestamp: time.Date(2025, 7, 14, 17, 52, 51, 0, time.UTC),
				Channel:   ChannelTrade,
				Player:    "SumoUndies",
				Body:      "any1 could help me get trough campaign starting at act 6 ? i can pay :)",
			},
			wantOK: true,
		},
		{
			name:  "party chat with guild",
			input: `2025/07/25 21:30:17 168597312 cff945b9 [INFO Client 59180] %<®Æå> 全身只剩下護甲這隻應該組的起來吧: TY`,
			want: Message{
				Timestamp: time.Date(2025, 7, 25, 21, 30, 17, 0, time.UTC),
				Channel:   ChannelParty,
				Guild:     "®Æå",
				Player:    "全身只剩下護甲這隻應該組的起來吧",
				Body:      "TY",
			},
			wantOK: true,
		},
		{
			name:  "japanese global chat",
			input: `2025/07/12 05:59:09 52180843 cff945b9 [INFO Client 51360] #弱毒性プチ: おはようごじあます`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 5, 59, 9, 0, time.UTC),
				Channel:   ChannelGlobal,
				Player:    "弱毒性プチ",
				Body:      "おはようごじあます",
			},
			wantOK: true,
		},
		{
			name:  "guild with special chars",
			input: `2025/07/12 23:20:23 114655390 cff945b9 [INFO Client 88228] @From <¿ZXC?™> Tiellonies: Hi, I would like to buy your map`,
			want: Message{
				Timestamp: time.Date(2025, 7, 12, 23, 20, 23, 0, time.UTC),
				Channel:   ChannelWhisperIn,
				Guild:     "¿ZXC?™",
				Player:    "Tiellonies",
				Body:      "Hi, I would like to buy your map",
			},
			wantOK: true,
		},
		{
			name:   "system message - not chat",
			input:  `2025/07/12 04:25:46 46578359 cff945b9 [INFO Client 51360] : You have entered Karui Shores.`,
			wantOK: false,
		},
		{
			name:   "engine log - not chat",
			input:  `2025/07/12 04:25:03 46535421 f2498ec7 [INFO Client 51360] [JOB] Irrecoverable Exception Callback: SET`,
			wantOK: false,
		},
		{
			name:   "empty line",
			input:  "",
			wantOK: false,
		},
		{
			name:   "log file opening",
			input:  `2025/07/12 04:25:03 ***** LOG FILE OPENING *****`,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseLine(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ParseLine() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if got.Timestamp != tt.want.Timestamp {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
			}
			if got.Channel != tt.want.Channel {
				t.Errorf("Channel = %v, want %v", got.Channel, tt.want.Channel)
			}
			if got.Guild != tt.want.Guild {
				t.Errorf("Guild = %q, want %q", got.Guild, tt.want.Guild)
			}
			if got.Player != tt.want.Player {
				t.Errorf("Player = %q, want %q", got.Player, tt.want.Player)
			}
			if got.Body != tt.want.Body {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}
