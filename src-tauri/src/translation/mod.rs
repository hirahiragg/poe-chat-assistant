pub mod cache;
pub mod deepl;
pub mod gemini;
pub mod google;
pub mod service;

use std::sync::LazyLock;

use serde::{Deserialize, Serialize};

pub use service::TranslationService;

static HTTP_CLIENT: LazyLock<reqwest::Client> = LazyLock::new(reqwest::Client::new);

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum Direction {
    Inbound,
    Outbound,
}

#[derive(Debug, Clone)]
pub struct TranslationRequest {
    pub direction: Direction,
    pub message: String,
    pub target_lang: String,
    pub context_player: String,
    pub context_body: String,
}

#[derive(Debug, Clone)]
pub enum TranslatorKind {
    Google,
    DeepL(String),
    Gemini(String),
}

/// Dispatch a translation request to the appropriate backend.
pub async fn translate(kind: &TranslatorKind, req: &TranslationRequest) -> Result<String, String> {
    match kind {
        TranslatorKind::Google => google::translate(&HTTP_CLIENT, req).await,
        TranslatorKind::DeepL(api_key) => deepl::translate(&HTTP_CLIENT, api_key, req).await,
        TranslatorKind::Gemini(api_key) => gemini::translate(&HTTP_CLIENT, api_key, req).await,
    }
}
