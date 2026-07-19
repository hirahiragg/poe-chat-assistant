package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	LogPath    string `json:"log_path"`
	Translator string `json:"translator"`
	DeepLKey   string `json:"deepl_api_key,omitempty"`
	GeminiKey  string `json:"gemini_api_key,omitempty"`
	SourceLang string `json:"source_language"`
	TargetLang string `json:"target_language"`
}

func Default() *Config {
	return &Config{
		Translator: "google",
		SourceLang: "en",
		TargetLang: "ja",
	}
}

func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "poe-chat-assistant"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() *Config {
	cfg := Default()
	p, err := Path()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, cfg)
	if cfg.Translator == "" {
		cfg.Translator = "google"
	}
	if cfg.SourceLang == "" {
		cfg.SourceLang = "en"
	}
	if cfg.TargetLang == "" {
		cfg.TargetLang = "ja"
	}
	return cfg
}

func (c *Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func (c *Config) APIKey() string {
	switch c.Translator {
	case "deepl":
		return c.DeepLKey
	case "gemini":
		return c.GeminiKey
	default:
		return ""
	}
}
