import { useCallback, useEffect, useRef, useState } from "react";
import { open } from "@tauri-apps/plugin-dialog";
import { DEFAULT_CONFIG, type Config } from "../types/config";

interface SettingsOverlayProps {
  config: Config;
  onSave: (config: Config) => Promise<void>;
  onClose: () => void;
}

const translators = ["Google", "DeepL", "Gemini"];
const presetLangs = ["en", "ja", "ko", "zh"];

export default function SettingsOverlay({
  config,
  onSave,
  onClose,
}: SettingsOverlayProps) {
  const [draft, setDraft] = useState<Config>({ ...config });
  const [recording, setRecording] = useState(false);
  const [confirmReset, setConfirmReset] = useState(false);
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const autoSave = useCallback(
    (newDraft: Config) => {
      if (saveTimer.current) clearTimeout(saveTimer.current);
      saveTimer.current = setTimeout(() => {
        onSave(newDraft);
      }, 500);
    },
    [onSave],
  );

  useEffect(() => {
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current);
    };
  }, []);

  const update = <K extends keyof Config>(key: K, value: Config[K]) => {
    const newDraft = { ...draft, [key]: value };
    setDraft(newDraft);
    autoSave(newDraft);
  };

  const keyToString = useCallback((e: KeyboardEvent): string | null => {
    const modifierKeys = new Set(["Control", "Shift", "Alt", "Meta"]);
    if (modifierKeys.has(e.key)) return null;

    const parts: string[] = [];
    if (e.ctrlKey || e.metaKey) parts.push("ctrl");
    if (e.shiftKey) parts.push("shift");
    if (e.altKey) parts.push("alt");

    const keyMap: Record<string, string> = {
      " ": "space",
      Enter: "enter",
      Tab: "tab",
      Backspace: "backspace",
      Escape: "escape",
    };

    const key =
      keyMap[e.key] ??
      (e.key.length === 1 ? e.key.toLowerCase() : e.key.toLowerCase());
    parts.push(key);
    return parts.join("+");
  }, []);

  useEffect(() => {
    if (!recording) return;
    const handler = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (e.key === "Escape") {
        setRecording(false);
        return;
      }
      const combo = keyToString(e);
      if (combo) {
        update("hotkey", combo);
        setRecording(false);
      }
    };
    window.addEventListener("keydown", handler, true);
    return () => window.removeEventListener("keydown", handler, true);
  }, [recording, keyToString]);

  const handleBrowse = async () => {
    const selected = await open({
      filters: [{ name: "Text", extensions: ["txt"] }],
    });
    if (selected) {
      update("log_path", selected as string);
    }
  };

  const handleReset = () => {
    const newDraft = { ...DEFAULT_CONFIG };
    setDraft(newDraft);
    setConfirmReset(false);
    onSave(newDraft);
  };

  const isPresetLang = presetLangs.includes(draft.target_language);

  return (
    <div className="h-screen w-screen bg-bg flex flex-col">
      {/* Header */}
      <div
        data-tauri-drag-region
        className="flex items-center justify-between px-4 py-3 border-b border-border"
      >
        <span className="text-text font-bold text-sm">Settings</span>
        <button
          onClick={onClose}
          className="text-text-dim text-xs hover:text-text transition-colors"
        >
          Close
        </button>
      </div>

      {/* Form */}
      <div className="flex-1 overflow-y-auto px-4 py-4 flex flex-col gap-4">
        {/* Client.txt Path */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Client.txt Path</label>
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={draft.log_path}
              onChange={(e) => update("log_path", e.target.value)}
              className="flex-1 bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim"
            />
            <button
              onClick={handleBrowse}
              className="bg-card text-text-dim hover:text-text rounded px-3 py-1.5 text-xs font-medium transition-colors flex-shrink-0"
            >
              Browse
            </button>
          </div>
        </div>

        {/* Translator */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Translator</label>
          <div className="flex gap-1">
            {translators.map((t) => (
              <button
                key={t}
                onClick={() => update("translator", t)}
                className={`rounded px-3 py-1.5 text-xs font-medium transition-colors ${
                  draft.translator === t
                    ? "bg-accent text-white"
                    : "bg-card text-text-dim hover:text-text"
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        {/* DeepL API Key */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">DeepL API Key</label>
          <input
            type="password"
            value={draft.deepl_api_key}
            onChange={(e) => update("deepl_api_key", e.target.value)}
            className="bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim"
          />
        </div>

        {/* Gemini API Key */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Gemini API Key</label>
          <input
            type="password"
            value={draft.gemini_api_key}
            onChange={(e) => update("gemini_api_key", e.target.value)}
            className="bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim"
          />
        </div>

        {/* Gemini Model */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Gemini Model</label>
          <input
            type="text"
            value={draft.gemini_model}
            onChange={(e) => update("gemini_model", e.target.value)}
            placeholder="gemini-3.5-flash"
            className="bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim"
          />
        </div>

        {/* Target Language */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Target Language</label>
          <div className="flex items-center gap-1">
            {presetLangs.map((lang) => (
              <button
                key={lang}
                onClick={() => update("target_language", lang)}
                className={`rounded px-3 py-1.5 text-xs font-medium transition-colors ${
                  draft.target_language === lang
                    ? "bg-accent text-white"
                    : "bg-card text-text-dim hover:text-text"
                }`}
              >
                {lang}
              </button>
            ))}
            <input
              type="text"
              value={isPresetLang ? "" : draft.target_language}
              onChange={(e) => update("target_language", e.target.value)}
              placeholder="custom"
              className="bg-card text-text border-none outline-none rounded px-3 py-1.5 text-xs w-16 text-center placeholder:text-text-dim"
            />
          </div>
        </div>

        {/* Hotkey */}
        <div className="flex flex-col gap-1">
          <label className="text-text-dim text-[11px]">Toggle Hotkey</label>
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={recording ? "Press a key combo..." : draft.hotkey}
              readOnly
              placeholder="e.g. ctrl+shift+space"
              className={`flex-1 bg-card text-text border-none outline-none rounded px-3 py-2 text-sm placeholder:text-text-dim ${
                recording ? "ring-1 ring-accent text-accent" : ""
              }`}
            />
            <button
              onClick={() => setRecording(!recording)}
              className={`rounded px-3 py-1.5 text-xs font-medium transition-colors flex-shrink-0 ${
                recording
                  ? "bg-red-600 text-white"
                  : "bg-card text-text-dim hover:text-text"
              }`}
            >
              {recording ? "Cancel" : "Record"}
            </button>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-end px-4 py-3 border-t border-border">
        {confirmReset ? (
          <div className="flex items-center gap-2">
            <span className="text-red-400 text-xs">Reset all settings?</span>
            <button
              onClick={handleReset}
              className="bg-red-600 text-white rounded px-3 py-1.5 text-xs font-medium hover:brightness-110 transition-colors"
            >
              Yes
            </button>
            <button
              onClick={() => setConfirmReset(false)}
              className="bg-card text-text-dim rounded px-3 py-1.5 text-xs font-medium hover:text-text transition-colors"
            >
              No
            </button>
          </div>
        ) : (
          <button
            onClick={() => setConfirmReset(true)}
            className="bg-red-900/50 text-red-400 rounded px-4 py-1.5 text-xs font-medium hover:bg-red-900/80 transition-colors"
          >
            Reset
          </button>
        )}
      </div>
    </div>
  );
}
