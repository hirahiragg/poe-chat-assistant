import { describe, expect, it } from "vitest";
import { Channel, type Message } from "../types/chat";
import type { ChannelFilters } from "../types/config";
import { filterMessages } from "./filter";

function msg(channel: Channel): Message {
  return {
    timestamp: "2025-07-12T04:26:51",
    channel,
    guild: "",
    player: "TestPlayer",
    body: "hello",
  };
}

const allMessages: Message[] = [
  msg(Channel.Global),
  msg(Channel.Trade),
  msg(Channel.Party),
  msg(Channel.Guild),
  msg(Channel.WhisperIn),
  msg(Channel.WhisperOut),
];

describe("filterMessages", () => {
  it("returns all messages when filters are empty (defaults to true)", () => {
    expect(filterMessages(allMessages, {})).toHaveLength(6);
  });

  it("filters out global when global=false", () => {
    const filters: ChannelFilters = { global: false };
    const result = filterMessages(allMessages, filters);
    expect(result).toHaveLength(5);
    expect(result.every((m) => m.channel !== Channel.Global)).toBe(true);
  });

  it("filters out whisper removes both in and out", () => {
    const filters: ChannelFilters = { whisper: false };
    const result = filterMessages(allMessages, filters);
    expect(result).toHaveLength(4);
    expect(
      result.every(
        (m) =>
          m.channel !== Channel.WhisperIn && m.channel !== Channel.WhisperOut,
      ),
    ).toBe(true);
  });

  it("filters out multiple channels", () => {
    const filters: ChannelFilters = { global: false, trade: false };
    const result = filterMessages(allMessages, filters);
    expect(result).toHaveLength(4);
  });

  it("returns empty array when all channels are disabled", () => {
    const filters: ChannelFilters = {
      global: false,
      trade: false,
      party: false,
      guild: false,
      whisper: false,
    };
    expect(filterMessages(allMessages, filters)).toHaveLength(0);
  });

  it("returns empty array for empty messages", () => {
    expect(filterMessages([], {})).toHaveLength(0);
  });
});
