import { useCallback, useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import type { Config } from "../types/config";

export function useConfig() {
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    invoke<Config>("get_config")
      .then((cfg) => {
        setConfig(cfg);
      })
      .catch((err) => {
        console.error("Failed to get config:", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  const saveConfig = useCallback(async (newConfig: Config) => {
    try {
      await invoke("save_config", { config: newConfig });
      setConfig(newConfig);
    } catch (err) {
      console.error("Failed to save config:", err);
      throw err;
    }
  }, []);

  return { config, saveConfig, loading };
}
