use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

#[cfg(target_os = "macos")]
extern "C" {
    fn foreground_app_name() -> *mut std::ffi::c_char;
    fn is_self_foreground() -> i32;
}

#[cfg(target_os = "macos")]
fn get_foreground_app() -> String {
    unsafe {
        let ptr = foreground_app_name();
        if ptr.is_null() {
            return String::new();
        }
        let s = std::ffi::CStr::from_ptr(ptr)
            .to_string_lossy()
            .into_owned();
        libc::free(ptr as *mut std::ffi::c_void);
        s
    }
}

#[cfg(target_os = "macos")]
fn is_self_focused() -> bool {
    unsafe { is_self_foreground() == 1 }
}

#[cfg(not(target_os = "macos"))]
fn get_foreground_app() -> String {
    String::new()
}

#[cfg(not(target_os = "macos"))]
fn is_self_focused() -> bool {
    false
}

const GAME_TITLE: &str = "Path of Exile";

pub fn is_poe_or_self_active() -> bool {
    if is_self_focused() {
        return true;
    }
    let name = get_foreground_app();
    name.to_lowercase().contains(&GAME_TITLE.to_lowercase())
}

pub fn start_focus_watcher(
    app_handle: tauri::AppHandle,
    cancel: Arc<AtomicBool>,
) {
    std::thread::spawn(move || {
        let mut was_active = false;

        while !cancel.load(Ordering::Relaxed) {
            let is_active = is_poe_or_self_active();

            if is_active != was_active {
                was_active = is_active;

                use tauri::Manager;
                let state = app_handle.state::<crate::state::AppState>();
                let hotkey = state.config.lock().unwrap().hotkey.clone();

                if is_active {
                    crate::register_hotkey(&app_handle, "", &hotkey);
                } else {
                    crate::register_hotkey(&app_handle, &hotkey, "");
                }
            }

            std::thread::sleep(std::time::Duration::from_secs(1));
        }
    });
}
