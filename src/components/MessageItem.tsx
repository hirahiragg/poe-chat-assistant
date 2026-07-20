import { channelColor, channelSymbol, type Message } from "../types/chat";

interface MessageItemProps {
  message: Message;
  selected: boolean;
  onClick: () => void;
}

const shortFmt = new Intl.DateTimeFormat(undefined, {
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

function formatShortTimestamp(ts: string): string {
  const d = new Date(ts);
  if (isNaN(d.getTime())) return ts;
  return shortFmt.format(d);
}

function truncate(text: string, max: number): string {
  if (text.length <= max) return text;
  return text.slice(0, max) + "...";
}

export default function MessageItem({
  message,
  selected,
  onClick,
}: MessageItemProps) {
  const symbol = channelSymbol(message.channel);
  const colorClass = channelColor(message.channel);

  return (
    <div
      onClick={onClick}
      className={`px-3 py-2 cursor-pointer transition-colors ${
        selected ? "bg-selected" : "hover:bg-card/50"
      }`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5 min-w-0">
          <span className={`${colorClass} font-bold text-xs flex-shrink-0`}>
            {symbol}
          </span>
          <span className="text-white font-bold text-xs truncate">
            {message.player}
          </span>
        </div>
        <span className="text-text-dim text-[11px] flex-shrink-0 ml-2 whitespace-nowrap">
          {formatShortTimestamp(message.timestamp)}
        </span>
      </div>
      <div className="text-text-dim text-[11px] mt-0.5 truncate">
        {truncate(message.body, 60)}
      </div>
    </div>
  );
}
