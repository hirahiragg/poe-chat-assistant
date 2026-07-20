import { describe, expect, it } from "vitest";
import { Channel, type Message } from "../types/chat";
import { cacheKey } from "./useTranslation";

function msg(overrides: Partial<Message> = {}): Message {
  return {
    timestamp: "2025-07-12T04:26:51",
    channel: Channel.Global,
    guild: "",
    player: "TestPlayer",
    body: "hello world",
    ...overrides,
  };
}

describe("cacheKey", () => {
  it("generates a deterministic key", () => {
    const m = msg();
    expect(cacheKey(m)).toBe(cacheKey(m));
  });

  it("includes all identifying fields", () => {
    const a = msg({ player: "Alice" });
    const b = msg({ player: "Bob" });
    expect(cacheKey(a)).not.toBe(cacheKey(b));
  });

  it("differs by channel", () => {
    const a = msg({ channel: Channel.Global });
    const b = msg({ channel: Channel.Trade });
    expect(cacheKey(a)).not.toBe(cacheKey(b));
  });

  it("differs by body", () => {
    const a = msg({ body: "hello" });
    const b = msg({ body: "world" });
    expect(cacheKey(a)).not.toBe(cacheKey(b));
  });

  it("differs by timestamp", () => {
    const a = msg({ timestamp: "2025-07-12T04:26:51" });
    const b = msg({ timestamp: "2025-07-12T04:26:52" });
    expect(cacheKey(a)).not.toBe(cacheKey(b));
  });
});
