import { useCallback, useEffect, useRef, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import type { Message } from "../types/chat";

export type LoadState = "idle" | "loading" | "eof" | "not_found";

export function useMessages() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const unlistenRef = useRef<(() => void) | null>(null);

  const fetchMessages = useCallback(async () => {
    try {
      const msgs = await invoke<Message[]>("get_messages");
      setMessages(msgs);
    } catch (err) {
      console.error("Failed to get messages:", err);
    }
  }, []);

  useEffect(() => {
    fetchMessages();

    let cancelled = false;
    listen<void>("new-message", () => {
      if (!cancelled) {
        fetchMessages();
      }
    }).then((unlisten) => {
      if (cancelled) {
        unlisten();
      } else {
        unlistenRef.current = unlisten;
      }
    });

    return () => {
      cancelled = true;
      if (unlistenRef.current) {
        unlistenRef.current();
        unlistenRef.current = null;
      }
    };
  }, [fetchMessages]);

  const loadMore = useCallback(async () => {
    if (loadState === "loading" || loadState === "eof") return;
    setLoadState("loading");
    try {
      const result = await invoke<string>("load_more");
      if (result === "eof") {
        setLoadState("eof");
      } else if (result === "not_found") {
        setLoadState("not_found");
      } else {
        setLoadState("idle");
      }
      await fetchMessages();
    } catch (err) {
      console.error("Failed to load more:", err);
      setLoadState("not_found");
    }
  }, [loadState, fetchMessages]);

  return { messages, loadMore, loadState };
}
