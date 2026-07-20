export enum Channel {
  Global = "Global",
  Trade = "Trade",
  Party = "Party",
  Guild = "Guild",
  WhisperIn = "WhisperIn",
  WhisperOut = "WhisperOut",
}

export interface Message {
  timestamp: string;
  channel: Channel;
  guild: string;
  player: string;
  body: string;
}

export function channelSymbol(ch: Channel): string {
  switch (ch) {
    case Channel.Global:
      return "#";
    case Channel.Trade:
      return "$";
    case Channel.Party:
      return "%";
    case Channel.Guild:
      return "&";
    case Channel.WhisperIn:
      return "@";
    case Channel.WhisperOut:
      return "->";
  }
}

export function channelColor(ch: Channel): string {
  switch (ch) {
    case Channel.Global:
      return "text-ch-global";
    case Channel.Trade:
      return "text-ch-trade";
    case Channel.Party:
      return "text-ch-party";
    case Channel.Guild:
      return "text-ch-guild";
    case Channel.WhisperIn:
    case Channel.WhisperOut:
      return "text-ch-whisper";
  }
}

export function channelBgColor(ch: Channel): string {
  switch (ch) {
    case Channel.Global:
      return "bg-ch-global";
    case Channel.Trade:
      return "bg-ch-trade";
    case Channel.Party:
      return "bg-ch-party";
    case Channel.Guild:
      return "bg-ch-guild";
    case Channel.WhisperIn:
    case Channel.WhisperOut:
      return "bg-ch-whisper";
  }
}
