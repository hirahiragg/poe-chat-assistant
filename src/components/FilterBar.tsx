import type { ChannelFilters } from "../types/config";

interface FilterBarProps {
  filters: ChannelFilters;
  onToggle: (key: keyof ChannelFilters) => void;
  onSettingsClick: () => void;
}

const filterButtons: {
  key: keyof ChannelFilters;
  symbol: string;
  activeClass: string;
}[] = [
  { key: "global", symbol: "#", activeClass: "bg-ch-global" },
  { key: "whisper", symbol: "@", activeClass: "bg-ch-whisper" },
  { key: "guild", symbol: "&", activeClass: "bg-ch-guild" },
  { key: "party", symbol: "%", activeClass: "bg-ch-party" },
  { key: "trade", symbol: "$", activeClass: "bg-ch-trade" },
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
        {filterButtons.map(({ key, symbol, activeClass }) => {
          const active = filters[key] !== false;
          return (
            <button
              key={key}
              onClick={() => onToggle(key)}
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
