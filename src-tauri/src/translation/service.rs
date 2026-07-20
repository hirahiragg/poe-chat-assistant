use super::cache::Cache;
use super::TranslatorKind;

pub struct TranslationService {
    pub kind: TranslatorKind,
    pub cache: Cache,
}

impl TranslationService {
    pub fn new(config: &crate::config::Config) -> Self {
        let kind = match config.translator.to_lowercase().as_str() {
            "deepl" => TranslatorKind::DeepL(config.deepl_api_key.clone()),
            "gemini" => TranslatorKind::Gemini(config.gemini_api_key.clone()),
            _ => TranslatorKind::Google,
        };
        Self {
            kind,
            cache: Cache::new(),
        }
    }
}
