use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

// --- macOS: Objective-C via FFI (macos.m) ---

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

// --- Windows: Win32 API via raw FFI ---

#[cfg(target_os = "windows")]
#[link(name = "user32")]
extern "system" {
    fn GetForegroundWindow() -> isize;
    fn GetWindowTextW(hwnd: isize, string: *mut u16, max_count: i32) -> i32;
    fn GetWindowThreadProcessId(hwnd: isize, process_id: *mut u32) -> u32;
}

#[cfg(target_os = "windows")]
#[link(name = "kernel32")]
extern "system" {
    fn GetCurrentProcessId() -> u32;
}

#[cfg(target_os = "windows")]
fn get_foreground_app() -> String {
    unsafe {
        let hwnd = GetForegroundWindow();
        if hwnd == 0 {
            return String::new();
        }
        let mut buf = [0u16; 256];
        let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), buf.len() as i32);
        if len <= 0 {
            return String::new();
        }
        String::from_utf16_lossy(&buf[..len as usize])
    }
}

#[cfg(target_os = "windows")]
fn is_self_focused() -> bool {
    unsafe {
        let hwnd = GetForegroundWindow();
        if hwnd == 0 {
            return false;
        }
        let mut pid: u32 = 0;
        GetWindowThreadProcessId(hwnd, &mut pid);
        let our_pid = GetCurrentProcessId();
        pid == our_pid
    }
}

// --- Linux / other: stub ---

#[cfg(not(any(target_os = "macos", target_os = "windows")))]
fn get_foreground_app() -> String {
    String::new()
}

#[cfg(not(any(target_os = "macos", target_os = "windows")))]
fn is_self_focused() -> bool {
    false
}

const GAME_TITLES: &[&str] = &["path of exile"];

pub fn is_poe_or_self_active() -> bool {
    if is_self_focused() {
        return true;
    }
    let name = get_foreground_app().to_lowercase();
    GAME_TITLES.iter().any(|t| name.contains(t))
}

pub fn start_focus_watcher(
    app_handle: tauri::AppHandle,
    cancel: Arc<AtomicBool>,
) {
    std::thread::spawn(move || {
        let mut was_active = false;
        let mut was_self = false;

        while !cancel.load(Ordering::Relaxed) {
            let is_active = is_poe_or_self_active();
            let is_self = is_self_focused();

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

            if was_self && !is_self {
                if let Some(window) = app_handle.get_webview_window("main") {
                    let _ = window.hide();
                }
            }

            was_self = is_self;
            std::thread::sleep(std::time::Duration::from_millis(250));
        }
    });
}
