use std::sync::atomic::AtomicBool;
use std::sync::Arc;

use tauri::State;

use crate::chat::Message;
use crate::config::Config;
use crate::state::AppState;
use crate::translation::{Direction, TranslationRequest};

#[tauri::command]
pub async fn get_messages(state: State<'_, AppState>) -> Result<Vec<Message>, String> {
    let store = state.store.lock().unwrap();
    Ok(store.list())
}

#[tauri::command]
pub async fn get_config(state: State<'_, AppState>) -> Result<Config, String> {
    let config = state.config.lock().unwrap();
    Ok(config.clone())
}

#[tauri::command]
pub async fn save_config(
    app: tauri::AppHandle,
    state: State<'_, AppState>,
    config: Config,
) -> Result<(), String> {
    // Save config to disk
    config.save()?;

    // Get old hotkey before updating config
    let old_hotkey = state.config.lock().unwrap().hotkey.clone();

    // Stop old watcher
    {
        let cancel = state.watcher_cancel.lock().unwrap();
        cancel.store(true, std::sync::atomic::Ordering::Relaxed);
    }

    // Update config in state
    {
        let mut cfg = state.config.lock().unwrap();
        *cfg = config.clone();
    }

    // Update translation service
    {
        let mut service = state.service.lock().unwrap();
        *service = crate::translation::TranslationService::new(&config);
    }

    // Reset load_more offset
    {
        let mut offset = state.load_more_offset.lock().unwrap();
        *offset = 0;
    }

    // Reset the message store
    {
        let mut store = state.store.lock().unwrap();
        *store = crate::chat::Store::new(500);
    }

    // Create a new cancel token for the new watcher
    let new_cancel = Arc::new(AtomicBool::new(false));
    {
        let mut cancel = state.watcher_cancel.lock().unwrap();
        *cancel = new_cancel;
    }

    // Start new watcher (reads initial tail and spawns watch thread)
    crate::start_watcher(&app);

    // Re-register hotkey
    crate::register_hotkey(&app, &old_hotkey, &config.hotkey);

    Ok(())
}

#[tauri::command]
pub async fn translate_message(
    state: State<'_, AppState>,
    direction: Direction,
    message: String,
    target_lang: String,
    context_player: String,
    context_body: String,
) -> Result<String, String> {
    let req = TranslationRequest {
        direction,
        message,
        target_lang,
        context_player,
        context_body,
    };

    // Check cache first (short lock)
    {
        let service = state.service.lock().unwrap();
        if let Some(cached) = service.cache.get(&req) {
            return Ok(cached);
        }
    }

    // Clone the translator kind so we can release the lock
    let kind = {
        let service = state.service.lock().unwrap();
        service.kind.clone()
    };

    // Perform translation without holding any locks
    let result = crate::translation::translate(&kind, &req).await?;

    // Cache the result (short lock)
    {
        let mut service = state.service.lock().unwrap();
        service.cache.set(&req, &result);
    }

    Ok(result)
}

#[tauri::command]
pub async fn load_more(state: State<'_, AppState>) -> Result<String, String> {
    let cfg = state.config.lock().unwrap().clone();
    if cfg.log_path.is_empty() {
        return Ok("eof".to_string());
    }

    let watcher = crate::logwatcher::Watcher::new(&cfg.log_path);
    let chunk_size: i64 = 512 * 1024;

    for _ in 0..20 {
        let current_offset = {
            let mut offset = state.load_more_offset.lock().unwrap();
            let current = *offset;
            *offset += chunk_size;
            current
        };

        let lines = match watcher.read_range(current_offset, chunk_size) {
            Ok(lines) => lines,
            Err(_) => return Ok("eof".to_string()),
        };

        if lines.is_empty() {
            return Ok("eof".to_string());
        }

        let mut new_msgs = Vec::new();
        for line in &lines {
            if let Some(msg) = crate::chat::parser::parse_line(line) {
                new_msgs.push(msg);
            }
        }

        if !new_msgs.is_empty() {
            let mut store = state.store.lock().unwrap();
            store.prepend(new_msgs);
            return Ok("found".to_string());
        }
    }

    Ok("not_found".to_string())
}
