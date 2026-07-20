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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::translation::Direction;

    fn req(msg: &str, direction: Direction) -> TranslationRequest {
        TranslationRequest {
            direction,
            message: msg.to_string(),
            target_lang: "ja".to_string(),
            context_player: String::new(),
            context_body: String::new(),
        }
    }

    #[test]
    fn cache_hit() {
        let mut cache = Cache::new();
        let r = req("hello", Direction::Inbound);
        cache.set(&r, "こんにちは");
        assert_eq!(cache.get(&r), Some("こんにちは".to_string()));
    }

    #[test]
    fn cache_miss() {
        let cache = Cache::new();
        let r = req("hello", Direction::Inbound);
        assert_eq!(cache.get(&r), None);
    }

    #[test]
    fn different_messages_different_keys() {
        let mut cache = Cache::new();
        let r1 = req("hello", Direction::Inbound);
        let r2 = req("world", Direction::Inbound);
        cache.set(&r1, "result1");
        assert_eq!(cache.get(&r1), Some("result1".to_string()));
        assert_eq!(cache.get(&r2), None);
    }

    #[test]
    fn different_direction_different_keys() {
        let mut cache = Cache::new();
        let r1 = req("hello", Direction::Inbound);
        let r2 = req("hello", Direction::Outbound);
        cache.set(&r1, "inbound_result");
        assert_eq!(cache.get(&r1), Some("inbound_result".to_string()));
        assert_eq!(cache.get(&r2), None);
    }

    #[test]
    fn overwrite_existing_entry() {
        let mut cache = Cache::new();
        let r = req("hello", Direction::Inbound);
        cache.set(&r, "old");
        cache.set(&r, "new");
        assert_eq!(cache.get(&r), Some("new".to_string()));
    }

    #[test]
    fn key_is_deterministic() {
        let r = req("hello", Direction::Inbound);
        assert_eq!(make_key(&r), make_key(&r));
    }

    #[test]
    fn context_affects_key() {
        let r1 = TranslationRequest {
            direction: Direction::Outbound,
            message: "hello".to_string(),
            target_lang: "en".to_string(),
            context_player: "Alice".to_string(),
            context_body: "hi".to_string(),
        };
        let r2 = TranslationRequest {
            context_player: "Bob".to_string(),
            ..r1.clone()
        };
        assert_ne!(make_key(&r1), make_key(&r2));
    }
}
