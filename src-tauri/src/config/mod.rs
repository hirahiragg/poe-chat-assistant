use std::path::PathBuf;

use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, Default)]
pub struct ChannelFilters {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub global: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub whisper: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub guild: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub party: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub trade: Option<bool>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Config {
    pub log_path: String,
    pub translator: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub deepl_api_key: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub gemini_api_key: String,
    #[serde(default = "default_gemini_model", skip_serializing_if = "String::is_empty")]
    pub gemini_model: String,
    pub target_language: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub hotkey: String,
    #[serde(default)]
    pub channel_filters: ChannelFilters,
}

fn default_gemini_model() -> String {
    "gemini-3.5-flash".to_string()
}

impl Default for Config {
    fn default() -> Self {
        Self {
            log_path: r"C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs\Client.txt".to_string(),
            translator: "Google".to_string(),
            deepl_api_key: String::new(),
            gemini_api_key: String::new(),
            gemini_model: default_gemini_model(),
            target_language: "ja".to_string(),
            hotkey: String::new(),
            channel_filters: ChannelFilters::default(),
        }
    }
}

pub fn config_path() -> PathBuf {
    let base = dirs::config_dir().unwrap_or_else(|| PathBuf::from("."));
    base.join("poe-chat-assistant").join("config.json")
}

pub fn load() -> Config {
    let path = config_path();
    let data = match std::fs::read_to_string(&path) {
        Ok(d) => d,
        Err(_) => return Config::default(),
    };
    let mut cfg: Config = match serde_json::from_str(&data) {
        Ok(c) => c,
        Err(_) => return Config::default(),
    };
    if cfg.translator.is_empty() {
        cfg.translator = "Google".to_string();
    }
    if cfg.target_language.is_empty() {
        cfg.target_language = "ja".to_string();
    }
    cfg
}

impl Config {
    pub fn save(&self) -> Result<(), String> {
        let path = config_path();
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent).map_err(|e| format!("create config dir: {}", e))?;
        }
        let data =
            serde_json::to_string_pretty(self).map_err(|e| format!("serialize config: {}", e))?;
        std::fs::write(&path, data).map_err(|e| format!("write config: {}", e))?;
        Ok(())
    }
}
