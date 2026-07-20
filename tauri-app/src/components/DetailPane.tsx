import { useEffect, useState } from "react";
import {
  Channel,
  channelColor,
  channelSymbol,
  type Message,
} from "../types/chat";
import { cacheKey, useTranslation } from "../hooks/useTranslation";
import type { Config } from "../types/config";

interface DetailPaneProps {
  message: Message;
  config: Config;
}

function formatFullTimestamp(ts: string): string {
  const d = new Date(ts);
  if (isNaN(d.getTime())) return ts;
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const min = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return `${yyyy}/${mm}/${dd} ${hh}:${min}:${ss}`;
}

export default function DetailPane({ message, config }: DetailPaneProps) {
  const {
    translatingMsg,
    translatingOut,
    translateInbound,
    translateOutbound,
    getCache,
    updateReplyText,
  } = useTranslation();

  const key = cacheKey(message);
  const cached = getCache(key);
  const [replyText, setReplyText] = useState(cached?.replyText ?? "");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const entry = getCache(cacheKey(message));
    setReplyText(entry?.replyText ?? "");
  }, [message, getCache]);

  const handleReplyChange = (text: string) => {
    setReplyText(text);
    updateReplyText(key, text);
  };

  const handleTranslateInbound = () => {
    translateInbound(message, config.target_language);
  };

  const handleTranslateOutbound = () => {
    translateOutbound(
      message,
      config.target_language,
      message.player,
      message.body,
    );
  };

  const handleCopy = () => {
    if (cached?.translatedOut) {
      let text = cached.translatedOut;
      if (message.channel === Channel.WhisperIn || message.channel === Channel.WhisperOut) {
        text = `@${message.player} ${text}`;
      }
      navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  };

  const symbol = channelSymbol(message.channel);
  const colorClass = channelColor(message.channel);
  const hasInbound = !!cached?.translatedMsg;
  const hasOutbound = !!cached?.translatedOut;

  return (
    <div className="flex flex-col h-full p-4 overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className={`${colorClass} font-bold text-sm`}>{symbol}</span>
          <span className="text-white font-bold text-sm">{message.player}</span>
        </div>
        <span className="text-text-dim text-[11px]">
          {formatFullTimestamp(message.timestamp)}
        </span>
      </div>

      {/* Original message */}
      <div className="text-text text-sm select-text mb-3 leading-relaxed">
        {message.body}
      </div>

      {/* Translate inbound */}
      <div className="mb-3">
        <div className="flex justify-end">
          <button
            onClick={handleTranslateInbound}
            disabled={translatingMsg}
            className="bg-btn-bg text-btn-text rounded px-3 py-1.5 text-xs font-medium hover:brightness-110 transition-colors disabled:opacity-50"
          >
            {translatingMsg ? "..." : hasInbound ? "Re-translate" : "Translate"}
          </button>
        </div>
        {hasInbound && (
          <div className="text-translated text-sm mt-2 leading-relaxed">
            {cached!.translatedMsg}
          </div>
        )}
      </div>

      {/* Divider */}
      <div className="border-t border-border my-2" />

      {/* Reply section */}
      <div className="flex flex-col gap-2">
        <span className="text-text-dim text-[11px]">Reply</span>
        <input
          type="text"
          value={replyText}
          onChange={(e) => handleReplyChange(e.target.value)}
          placeholder="Type in Japanese..."
          className="bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim"
        />
        <div className="flex justify-end">
          <button
            onClick={handleTranslateOutbound}
            disabled={translatingOut || !replyText.trim()}
            className="bg-btn-bg text-btn-text rounded px-3 py-1.5 text-xs font-medium hover:brightness-110 transition-colors disabled:opacity-50"
          >
            {translatingOut ? "..." : "Translate"}
          </button>
        </div>
        {hasOutbound && (
          <div className="flex flex-col gap-1.5">
            <span className="text-text-dim text-[11px]">English</span>
            <div className="text-accent text-sm leading-relaxed">
              {cached!.translatedOut}
            </div>
            <div className="flex justify-end">
              <button
                onClick={handleCopy}
                className="bg-btn-bg text-btn-text rounded px-3 py-1.5 text-xs font-medium hover:brightness-110 transition-colors"
              >
                {copied ? "Copied!" : "Copy"}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
