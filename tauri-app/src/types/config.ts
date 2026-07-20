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
  target_language: string;
  hotkey: string;
  channel_filters: ChannelFilters;
}
