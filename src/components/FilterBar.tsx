import type { ChannelFilters } from "../types/config";

interface FilterBarProps {
  filters: ChannelFilters;
  onToggle: (key: keyof ChannelFilters) => void;
  onSettingsClick: () => void;
}

const filterButtons: {
  key: keyof ChannelFilters;
  symbol: string;
  label: string;
  activeClass: string;
}[] = [
  { key: "global", symbol: "#", label: "Global", activeClass: "bg-ch-global" },
  { key: "whisper", symbol: "@", label: "Whisper", activeClass: "bg-ch-whisper" },
  { key: "guild", symbol: "&", label: "Guild", activeClass: "bg-ch-guild" },
  { key: "party", symbol: "%", label: "Party", activeClass: "bg-ch-party" },
  { key: "trade", symbol: "$", label: "Trade", activeClass: "bg-ch-trade" },
];

export default function FilterBar({
  filters,
  onToggle,
  onSettingsClick,
}: FilterBarProps) {
  return (
    <div
      data-tauri-drag-region
      className="flex items-center justify-between px-2 py-1.5 bg-surface border-b border-border"
    >
      <div className="flex items-center gap-1">
        {filterButtons.map(({ key, symbol, label, activeClass }) => {
          const active = filters[key] !== false;
          return (
            <button
              key={key}
              onClick={() => onToggle(key)}
              title={label}
              className={`w-7 h-7 rounded text-xs font-bold flex items-center justify-center transition-colors ${
                active
                  ? `${activeClass} text-white`
                  : "bg-card text-text-dim"
              }`}
            >
              {symbol}
            </button>
          );
        })}
      </div>
      <button
        onClick={onSettingsClick}
        className="text-text-dim text-xs hover:text-text transition-colors"
      >
        Settings
      </button>
    </div>
  );
}
