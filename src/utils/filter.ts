import { Channel, type Message } from "../types/chat";
import type { ChannelFilters } from "../types/config";

export function filterMessages(
  messages: Message[],
  filters: ChannelFilters,
): Message[] {
  return messages.filter((msg) => {
    switch (msg.channel) {
      case Channel.Global:
        return filters.global !== false;
      case Channel.Trade:
        return filters.trade !== false;
      case Channel.Party:
        return filters.party !== false;
      case Channel.Guild:
        return filters.guild !== false;
      case Channel.WhisperIn:
      case Channel.WhisperOut:
        return filters.whisper !== false;
    }
  });
}
