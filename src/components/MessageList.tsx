import { useCallback, useLayoutEffect, useRef, useState } from "react";
import type { Message } from "../types/chat";
import type { LoadState } from "../hooks/useMessages";
import MessageItem from "./MessageItem";

interface MessageListProps {
  messages: Message[];
  selectedIndex: number | null;
  onSelect: (index: number) => void;
  onLoadMore: () => void;
  loadState: LoadState;
}

function loadButtonText(state: LoadState): string {
  switch (state) {
    case "idle":
      return "Load More";
    case "loading":
      return "Loading...";
    case "eof":
      return "No more messages";
    case "not_found":
      return "Load More (no chats found, retry)";
  }
}

const SCROLL_TOP_THRESHOLD = 30;

export default function MessageList({
  messages,
  selectedIndex,
  onSelect,
  onLoadMore,
  loadState,
}: MessageListProps) {
  const listRef = useRef<HTMLDivElement>(null);
  const prevCountRef = useRef(messages.length);
  const prevScrollHeightRef = useRef(0);
  const isLoadMore = useRef(false);
  const isNearTop = useRef(true);
  const [newCount, setNewCount] = useState(0);

  const handleScroll = useCallback(() => {
    if (!listRef.current) return;
    isNearTop.current = listRef.current.scrollTop <= SCROLL_TOP_THRESHOLD;
    if (isNearTop.current && newCount > 0) {
      setNewCount(0);
    }
  }, [newCount]);

  useLayoutEffect(() => {
    if (!listRef.current) return;

    const added = messages.length - prevCountRef.current;
    if (added > 0 && !isLoadMore.current) {
      if (isNearTop.current) {
        listRef.current.scrollTop = 0;
      } else {
        const delta = listRef.current.scrollHeight - prevScrollHeightRef.current;
        listRef.current.scrollTop += delta;
        setNewCount((c) => c + added);
      }
    }

    prevScrollHeightRef.current = listRef.current.scrollHeight;
    isLoadMore.current = false;
    prevCountRef.current = messages.length;
  }, [messages]);

  const handleLoadMore = () => {
    isLoadMore.current = true;
    onLoadMore();
  };

  const scrollToTop = () => {
    if (listRef.current) listRef.current.scrollTop = 0;
    setNewCount(0);
  };

  return (
    <div className="relative flex-1 min-h-0">
      {newCount > 0 && (
        <button
          onClick={scrollToTop}
          className="absolute top-2 left-1/2 -translate-x-1/2 z-10 bg-accent text-white text-xs font-medium px-3 py-1.5 rounded-full shadow-lg hover:brightness-110 transition-all flex items-center gap-1.5 whitespace-nowrap"
        >
          <span>&#8593;</span>
          {newCount} new message{newCount > 1 ? "s" : ""}
          <span
            onClick={(e) => { e.stopPropagation(); setNewCount(0); }}
            className="ml-0.5 opacity-70 hover:opacity-100"
          >
            &#10005;
          </span>
        </button>
      )}
      <div ref={listRef} onScroll={handleScroll} className="h-full overflow-y-auto">
        {messages.map((msg, i) => (
          <MessageItem
            key={`${msg.timestamp}-${msg.channel}-${msg.player}-${i}`}
            message={msg}
            selected={selectedIndex === i}
            onClick={() => onSelect(i)}
          />
        ))}
        <div className="flex justify-center py-3">
          <button
            onClick={handleLoadMore}
            disabled={loadState === "loading" || loadState === "eof"}
            className={`text-xs px-3 py-1.5 rounded transition-colors ${
              loadState === "eof"
                ? "text-text-dim cursor-default"
                : loadState === "loading"
                  ? "text-text-dim cursor-wait"
                  : "bg-btn-bg text-btn-text hover:brightness-110"
            }`}
          >
            {loadButtonText(loadState)}
          </button>
        </div>
      </div>
    </div>
  );
}
