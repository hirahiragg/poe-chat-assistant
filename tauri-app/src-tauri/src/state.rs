use std::sync::atomic::AtomicBool;
use std::sync::{Arc, Mutex};

use crate::chat::Store;
use crate::config::Config;
use crate::translation::TranslationService;

pub struct AppState {
    pub store: Mutex<Store>,
    pub config: Mutex<Config>,
    pub service: Mutex<TranslationService>,
    /// Cancel token for the current watcher thread.
    /// Wrapped in Mutex so it can be swapped out when restarting the watcher.
    pub watcher_cancel: Mutex<Arc<AtomicBool>>,
    /// Tracks how many bytes from the end of the log file have been read
    /// for the "load more" feature.
    pub load_more_offset: Mutex<i64>,
    /// Cancel token for the focus watcher thread.
    pub focus_cancel: Arc<AtomicBool>,
}
