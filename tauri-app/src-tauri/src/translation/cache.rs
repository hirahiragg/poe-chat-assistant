use std::collections::HashMap;

use sha2::{Digest, Sha256};

use super::TranslationRequest;

pub struct Cache {
    entries: HashMap<String, String>,
}

impl Cache {
    pub fn new() -> Self {
        Self {
            entries: HashMap::new(),
        }
    }

    pub fn get(&self, req: &TranslationRequest) -> Option<String> {
        let key = make_key(req);
        self.entries.get(&key).cloned()
    }

    pub fn set(&mut self, req: &TranslationRequest, result: &str) {
        let key = make_key(req);
        self.entries.insert(key, result.to_string());
    }
}

fn make_key(req: &TranslationRequest) -> String {
    let direction_str = match req.direction {
        super::Direction::Inbound => "inbound",
        super::Direction::Outbound => "outbound",
    };

    let mut hasher = Sha256::new();
    hasher.update(format!(
        "{}\0{}\0{}\0{}\0{}",
        direction_str, req.message, req.target_lang, req.context_player, req.context_body
    ));
    format!("{:x}", hasher.finalize())
}
