export interface ChannelFilters {
  global?: boolean;
  whisper?: boolean;
  guild?: boolean;
  party?: boolean;
  trade?: boolean;
}

export interface Config {
  log_path: string;
  translator: string;
  deepl_api_key: string;
  gemini_api_key: string;
  gemini_model: string;
  target_language: string;
  hotkey: string;
  channel_filters: ChannelFilters;
}

export const DEFAULT_CONFIG: Config = {
  log_path: String.raw`C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs\Client.txt`,
  translator: "google",
  deepl_api_key: "",
  gemini_api_key: "",
  gemini_model: "gemini-3.5-flash",
  target_language: "ja",
  hotkey: "",
  channel_filters: {},
};
