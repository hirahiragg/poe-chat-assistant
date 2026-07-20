import { useCallback, useRef, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import type { Message } from "../types/chat";

interface CacheEntry {
  translatedMsg?: string;
  translatedOut?: string;
  replyText?: string;
}

function cacheKey(msg: Message): string {
  return `${msg.timestamp}|${msg.channel}|${msg.player}|${msg.body}`;
}

export function useTranslation() {
  const cacheRef = useRef<Map<string, CacheEntry>>(new Map());
  const [, setTick] = useState(0);
  const [translatingMsg, setTranslatingMsg] = useState(false);
  const [translatingOut, setTranslatingOut] = useState(false);
  const [translateError, setTranslateError] = useState<string | null>(null);

  const forceUpdate = useCallback(() => {
    setTick((t) => t + 1);
  }, []);

  const getCache = useCallback((key: string): CacheEntry | undefined => {
    return cacheRef.current.get(key);
  }, []);

  const translateInbound = useCallback(
    async (message: Message, targetLang: string) => {
      const key = cacheKey(message);
      setTranslatingMsg(true);
      setTranslateError(null);
      try {
        const result = await invoke<string>("translate_message", {
          message: message.body,
          targetLang,
          direction: "inbound",
          contextPlayer: "",
          contextBody: "",
        });
        const entry = cacheRef.current.get(key) ?? {};
        entry.translatedMsg = result;
        cacheRef.current.set(key, entry);
        forceUpdate();
      } catch (err) {
        const msg = typeof err === "string" ? err : String(err);
        setTranslateError(msg);
        console.error("Inbound translation failed:", err);
      } finally {
        setTranslatingMsg(false);
      }
    },
    [forceUpdate],
  );

  const translateOutbound = useCallback(
    async (
      message: Message,
      targetLang: string,
      contextPlayer: string,
      contextBody: string,
    ) => {
      const key = cacheKey(message);
      const entry = cacheRef.current.get(key);
      const replyText = entry?.replyText ?? "";
      if (!replyText.trim()) return;

      setTranslatingOut(true);
      try {
        const result = await invoke<string>("translate_message", {
          message: replyText,
          targetLang,
          direction: "outbound",
          contextPlayer,
          contextBody,
        });
        const current = cacheRef.current.get(key) ?? {};
        current.translatedOut = result;
        cacheRef.current.set(key, current);
        forceUpdate();
      } catch (err) {
        const msg = typeof err === "string" ? err : String(err);
        setTranslateError(msg);
        console.error("Outbound translation failed:", err);
      } finally {
        setTranslatingOut(false);
      }
    },
    [forceUpdate],
  );

  const updateReplyText = useCallback(
    (key: string, text: string) => {
      const entry = cacheRef.current.get(key) ?? {};
      entry.replyText = text;
      cacheRef.current.set(key, entry);
      forceUpdate();
    },
    [forceUpdate],
  );

  return {
    cache: cacheRef.current,
    translatingMsg,
    translatingOut,
    translateError,
    translateInbound,
    translateOutbound,
    getCache,
    updateReplyText,
  };
}

export { cacheKey };
