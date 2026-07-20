mod chat;
mod commands;
mod config;
mod focus;
mod logwatcher;
mod state;
mod translation;

use std::sync::atomic::AtomicBool;
use std::sync::{Arc, Mutex};

use state::AppState;

pub fn run() {
    let cfg = config::load();

    let store = chat::Store::new(500);
    let service = translation::TranslationService::new(&cfg);
    let watcher_cancel = Arc::new(AtomicBool::new(false));
    let focus_cancel = Arc::new(AtomicBool::new(false));

    let app_state = AppState {
        store: Mutex::new(store),
        config: Mutex::new(cfg.clone()),
        service: Mutex::new(service),
        watcher_cancel: Mutex::new(watcher_cancel),
        load_more_offset: Mutex::new(0),
        focus_cancel: focus_cancel.clone(),
    };

    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .plugin(
            tauri_plugin_global_shortcut::Builder::new()
                .with_handler(|app, _shortcut, event| {
                    use tauri::Manager;
                    use tauri_plugin_global_shortcut::ShortcutState;

                    if event.state == ShortcutState::Pressed {
                        if let Some(window) = app.get_webview_window("main") {
                            let visible = window.is_visible().unwrap_or(false);
                            let focused = window.is_focused().unwrap_or(false);
                            if visible && focused {
                                let _ = window.hide();
                            } else {
                                let _ = window.show();
                                let _ = window.set_focus();
                            }
                        }
                    }
                })
                .build(),
        )
        .manage(app_state)
        .invoke_handler(tauri::generate_handler![
            commands::get_messages,
            commands::get_config,
            commands::save_config,
            commands::translate_message,
            commands::load_more,
        ])
        .setup(|app| {
            use tauri::Manager;
            use tauri::menu::{Menu, MenuItem};
            use tauri::tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent};

            let handle = app.handle().clone();

            // System tray
            let show_i = MenuItem::with_id(app, "show", "Show", true, None::<&str>)?;
            let quit_i = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &quit_i])?;

            TrayIconBuilder::new()
                .icon(app.default_window_icon().cloned().unwrap())
                .menu(&menu)
                .on_menu_event(|app, event| {
                    use tauri::Manager;
                    match event.id.as_ref() {
                        "show" => {
                            if let Some(window) = app.get_webview_window("main") {
                                let _ = window.show();
                                let _ = window.set_focus();
                            }
                        }
                        "quit" => {
                            app.exit(0);
                        }
                        _ => {}
                    }
                })
                .on_tray_icon_event(|tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        use tauri::Manager;
                        let app = tray.app_handle();
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                })
                .build(app)?;

            // Start the log watcher (reads initial tail + spawns watch thread)
            start_watcher(&handle);

            // Start focus watcher (registers hotkey only when PoE or self is active)
            let state = app.state::<AppState>();
            focus::start_focus_watcher(handle.clone(), state.focus_cancel.clone());

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

/// Start watching the log file for new chat messages.
/// Reads the tail of the file to populate initial messages,
/// then spawns a thread to watch for new lines.
pub(crate) fn start_watcher(app_handle: &tauri::AppHandle) {
    use tauri::Manager;

    let state = app_handle.state::<AppState>();
    let cfg = state.config.lock().unwrap().clone();

    if cfg.log_path.is_empty() {
        return;
    }

    let watcher = logwatcher::Watcher::new(&cfg.log_path);
    let initial_chunk: i64 = 512 * 1024;

    // Read the tail of the log file for initial messages
    match watcher.read_tail(initial_chunk) {
        Ok(lines) => {
            let mut store = state.store.lock().unwrap();
            for line in &lines {
                if let Some(msg) = chat::parser::parse_line(line) {
                    store.add(msg);
                }
            }
        }
        Err(_) => {}
    }

    // Set load_more offset so subsequent loads go further back
    *state.load_more_offset.lock().unwrap() = initial_chunk;

    // Clone the cancel token for the watch thread
    let cancel = state.watcher_cancel.lock().unwrap().clone();
    let handle = app_handle.clone();

    std::thread::spawn(move || {
        let _ = watcher.watch(cancel, |line| {
            use tauri::{Emitter, Manager};

            if let Some(msg) = chat::parser::parse_line(&line) {
                let st = handle.state::<AppState>();
                st.store.lock().unwrap().add(msg.clone());
                let _ = handle.emit("new-message", &msg);
            }
        });
    });
}

/// Unregister the old hotkey and register a new one.
pub(crate) fn register_hotkey(
    app_handle: &tauri::AppHandle,
    old_hotkey: &str,
    new_hotkey: &str,
) {
    #[cfg(desktop)]
    {
        use tauri_plugin_global_shortcut::GlobalShortcutExt;

        // Unregister old shortcut
        if !old_hotkey.is_empty() {
            if let Some(shortcut) = parse_hotkey(old_hotkey) {
                let _ = app_handle.global_shortcut().unregister(shortcut);
            }
        }

        // Register new shortcut
        if !new_hotkey.is_empty() {
            if let Some(shortcut) = parse_hotkey(new_hotkey) {
                let _ = app_handle.global_shortcut().register(shortcut);
            }
        }
    }
}

/// Parse a hotkey string like "ctrl+shift+space" into a `Shortcut`.
#[cfg(desktop)]
fn parse_hotkey(s: &str) -> Option<tauri_plugin_global_shortcut::Shortcut> {
    use tauri_plugin_global_shortcut::{Code, Modifiers, Shortcut};

    if s.is_empty() {
        return None;
    }

    let mut mods = Modifiers::empty();
    let mut code: Option<Code> = None;

    for part in s.split('+').map(|p| p.trim()) {
        let lower = part.to_lowercase();
        match lower.as_str() {
            "ctrl" | "control" => mods |= Modifiers::CONTROL,
            "shift" => mods |= Modifiers::SHIFT,
            "alt" => mods |= Modifiers::ALT,
            "super" | "meta" | "cmd" | "command" => mods |= Modifiers::SUPER,
            "space" => code = Some(Code::Space),
            "enter" | "return" => code = Some(Code::Enter),
            "tab" => code = Some(Code::Tab),
            "backspace" => code = Some(Code::Backspace),
            "escape" | "esc" => code = Some(Code::Escape),
            "f1" => code = Some(Code::F1),
            "f2" => code = Some(Code::F2),
            "f3" => code = Some(Code::F3),
            "f4" => code = Some(Code::F4),
            "f5" => code = Some(Code::F5),
            "f6" => code = Some(Code::F6),
            "f7" => code = Some(Code::F7),
            "f8" => code = Some(Code::F8),
            "f9" => code = Some(Code::F9),
            "f10" => code = Some(Code::F10),
            "f11" => code = Some(Code::F11),
            "f12" => code = Some(Code::F12),
            other => {
                if other.len() == 1 {
                    code = char_to_code(other.chars().next().unwrap());
                }
            }
        }
    }

    let mods_opt = if mods.is_empty() { None } else { Some(mods) };
    code.map(|c| Shortcut::new(mods_opt, c))
}

#[cfg(desktop)]
fn char_to_code(c: char) -> Option<tauri_plugin_global_shortcut::Code> {
    use tauri_plugin_global_shortcut::Code;

    Some(match c.to_ascii_lowercase() {
        'a' => Code::KeyA,
        'b' => Code::KeyB,
        'c' => Code::KeyC,
        'd' => Code::KeyD,
        'e' => Code::KeyE,
        'f' => Code::KeyF,
        'g' => Code::KeyG,
        'h' => Code::KeyH,
        'i' => Code::KeyI,
        'j' => Code::KeyJ,
        'k' => Code::KeyK,
        'l' => Code::KeyL,
        'm' => Code::KeyM,
        'n' => Code::KeyN,
        'o' => Code::KeyO,
        'p' => Code::KeyP,
        'q' => Code::KeyQ,
        'r' => Code::KeyR,
        's' => Code::KeyS,
        't' => Code::KeyT,
        'u' => Code::KeyU,
        'v' => Code::KeyV,
        'w' => Code::KeyW,
        'x' => Code::KeyX,
        'y' => Code::KeyY,
        'z' => Code::KeyZ,
        '0' => Code::Digit0,
        '1' => Code::Digit1,
        '2' => Code::Digit2,
        '3' => Code::Digit3,
        '4' => Code::Digit4,
        '5' => Code::Digit5,
        '6' => Code::Digit6,
        '7' => Code::Digit7,
        '8' => Code::Digit8,
        '9' => Code::Digit9,
        _ => return None,
    })
}
