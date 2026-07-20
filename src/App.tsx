import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useMessages } from "./hooks/useMessages";
import { useConfig } from "./hooks/useConfig";
import type { ChannelFilters } from "./types/config";
import FilterBar from "./components/FilterBar";
import MessageList from "./components/MessageList";
import DetailPane from "./components/DetailPane";
import SettingsOverlay from "./components/SettingsOverlay";
import { filterMessages } from "./utils/filter";

export default function App() {
  const { messages, loadMore, loadState, refresh } = useMessages();
  const { config, saveConfig: rawSaveConfig, loading } = useConfig();
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const [splitPercent, setSplitPercent] = useState(40);
  const dragging = useRef(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const saveConfig = useCallback(async (newConfig: Parameters<typeof rawSaveConfig>[0]) => {
    await rawSaveConfig(newConfig);
    await refresh();
  }, [rawSaveConfig, refresh]);

  const filters = config?.channel_filters ?? {};

  const filteredMessages = useMemo(
    () => filterMessages(messages, filters),
    [messages, filters],
  );

  const handleToggleFilter = useCallback(
    (key: keyof ChannelFilters) => {
      if (!config) return;
      const current = config.channel_filters[key] !== false;
      const newFilters = { ...config.channel_filters, [key]: !current };
      const newConfig = { ...config, channel_filters: newFilters };
      saveConfig(newConfig);
    },
    [config, saveConfig],
  );

  const handleSelect = useCallback((index: number) => {
    setSelectedIndex(index);
  }, []);

  const handleDragStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    dragging.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onMove = (ev: MouseEvent) => {
      if (!dragging.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const pct = ((ev.clientX - rect.left) / rect.width) * 100;
      setSplitPercent(Math.min(70, Math.max(20, pct)));
    };
    const onUp = () => {
      dragging.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    };
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
  }, []);

  // Keyboard navigation
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (showSettings) {
        if (e.key === "Escape") {
          setShowSettings(false);
        }
        return;
      }

      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((prev) => {
          if (prev === null) return 0;
          return Math.min(prev + 1, filteredMessages.length - 1);
        });
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((prev) => {
          if (prev === null) return 0;
          return Math.max(prev - 1, 0);
        });
      } else if (e.key === "Escape") {
        setSelectedIndex(null);
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [showSettings, filteredMessages.length]);

  if (loading || !config) {
    return (
      <div className="h-screen w-screen bg-bg flex items-center justify-center">
        <span className="text-text-dim text-sm">Loading...</span>
      </div>
    );
  }

  if (showSettings) {
    return (
      <SettingsOverlay
        config={config}
        onSave={saveConfig}
        onClose={() => setShowSettings(false)}
      />
    );
  }

  const selectedMessage =
    selectedIndex !== null ? filteredMessages[selectedIndex] ?? null : null;

  return (
    <div className="h-screen w-screen bg-bg flex flex-col">
      <FilterBar
        filters={filters}
        onToggle={handleToggleFilter}
        onSettingsClick={() => setShowSettings(true)}
      />
      <div ref={containerRef} className="flex flex-1 min-h-0">
        <div style={{ width: `${splitPercent}%` }} className="flex flex-col bg-surface">
          <MessageList
            messages={filteredMessages}
            selectedIndex={selectedIndex}
            onSelect={handleSelect}
            onLoadMore={loadMore}
            loadState={loadState}
          />
        </div>

        <div
          onMouseDown={handleDragStart}
          className="w-1 bg-border hover:bg-accent cursor-col-resize flex-shrink-0 transition-colors"
        />

        <div style={{ width: `${100 - splitPercent}%` }} className="flex flex-col bg-surface">
          {selectedMessage ? (
            <DetailPane message={selectedMessage} config={config} />
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <span className="text-text-dim text-sm">Select a chat</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
