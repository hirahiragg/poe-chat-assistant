import { describe, expect, it } from "vitest";
import {
  Channel,
  channelBgColor,
  channelColor,
  channelSymbol,
} from "./chat";

describe("channelSymbol", () => {
  it.each([
    [Channel.Global, "#"],
    [Channel.Trade, "$"],
    [Channel.Party, "%"],
    [Channel.Guild, "&"],
    [Channel.WhisperIn, "@"],
    [Channel.WhisperOut, "->"],
  ])("returns correct symbol for %s", (channel, expected) => {
    expect(channelSymbol(channel)).toBe(expected);
  });
});

describe("channelColor", () => {
  it("returns whisper color for both whisper directions", () => {
    expect(channelColor(Channel.WhisperIn)).toBe("text-ch-whisper");
    expect(channelColor(Channel.WhisperOut)).toBe("text-ch-whisper");
  });

  it("returns unique color per non-whisper channel", () => {
    const colors = [
      channelColor(Channel.Global),
      channelColor(Channel.Trade),
      channelColor(Channel.Party),
      channelColor(Channel.Guild),
    ];
    expect(new Set(colors).size).toBe(4);
  });
});

describe("channelBgColor", () => {
  it("returns bg class for each channel", () => {
    for (const ch of Object.values(Channel)) {
      const bg = channelBgColor(ch);
      expect(bg).toMatch(/^bg-ch-/);
    }
  });

  it("returns same bg for both whisper directions", () => {
    expect(channelBgColor(Channel.WhisperIn)).toBe(
      channelBgColor(Channel.WhisperOut),
    );
  });
});
