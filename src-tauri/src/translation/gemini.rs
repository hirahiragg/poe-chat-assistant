use serde::{Deserialize, Serialize};

use super::TranslationRequest;

const INBOUND_SYSTEM_PROMPT_TMPL: &str = r#"You are translating Path of Exile chat messages.

Translate the message into natural {language}.

Rules:
- Understand Path of Exile terminology and slang.
- Keep item names, skill names, and currency names in English when appropriate.
- Keep common PoE abbreviations (div, ex, chaos, etc.) when appropriate.
- Do not explain the translation.
- Return only the translated text."#;

const OUTBOUND_SYSTEM_PROMPT: &str = r#"You are helping a Path of Exile player reply to another player.

Convert the message into natural, concise English suitable for Path of Exile chat.

Rules:
- Preserve the user's intended meaning.
- Use natural PoE chat language.
- Keep it concise as it will be typed into game chat.
- Do not add information the user did not provide.
- Do not explain the translation.
- Return only the English message."#;

fn lang_display_name(code: &str) -> &str {
    match code {
        "ja" => "Japanese",
        "en" => "English",
        "ko" => "Korean",
        "zh" => "Chinese",
        _ => code,
    }
}

// --- Gemini API request/response types ---

#[derive(Serialize)]
struct GeminiRequest {
    system_instruction: GeminiContent,
    contents: Vec<GeminiContent>,
    #[serde(rename = "generationConfig")]
    generation_config: GenerationConfig,
}

#[derive(Serialize)]
struct GeminiContent {
    parts: Vec<GeminiPart>,
}

#[derive(Serialize)]
struct GeminiPart {
    text: String,
}

#[derive(Serialize)]
struct GenerationConfig {
    temperature: f32,
    #[serde(rename = "maxOutputTokens")]
    max_output_tokens: u32,
}

#[derive(Deserialize)]
struct GeminiResponse {
    candidates: Option<Vec<GeminiCandidate>>,
}

#[derive(Deserialize)]
struct GeminiCandidate {
    content: Option<GeminiResponseContent>,
}

#[derive(Deserialize)]
struct GeminiResponseContent {
    parts: Option<Vec<GeminiResponsePart>>,
}

#[derive(Deserialize)]
struct GeminiResponsePart {
    text: Option<String>,
}

// ---

pub async fn translate(
    client: &reqwest::Client,
    api_key: &str,
    req: &TranslationRequest,
) -> Result<String, String> {
    let system_prompt = if req.direction == super::Direction::Outbound {
        OUTBOUND_SYSTEM_PROMPT.to_string()
    } else {
        let name = lang_display_name(&req.target_lang);
        INBOUND_SYSTEM_PROMPT_TMPL.replace("{language}", name)
    };

    let user_message = if req.direction == super::Direction::Outbound
        && !req.context_player.is_empty()
    {
        build_outbound_prompt(&req.message, &req.context_player, &req.context_body)
    } else {
        req.message.clone()
    };

    let url = format!(
        "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key={}",
        api_key
    );

    let body = GeminiRequest {
        system_instruction: GeminiContent {
            parts: vec![GeminiPart {
                text: system_prompt,
            }],
        },
        contents: vec![GeminiContent {
            parts: vec![GeminiPart {
                text: user_message,
            }],
        }],
        generation_config: GenerationConfig {
            temperature: 0.3,
            max_output_tokens: 256,
        },
    };

    let resp = client
        .post(&url)
        .json(&body)
        .send()
        .await
        .map_err(|e| format!("gemini request: {}", e))?;

    let status = resp.status();
    let resp_body = resp
        .text()
        .await
        .map_err(|e| format!("gemini read: {}", e))?;

    if !status.is_success() {
        return Err(format!("gemini status {}: {}", status, resp_body));
    }

    let response: GeminiResponse =
        serde_json::from_str(&resp_body).map_err(|e| format!("gemini parse: {}", e))?;

    extract_text(&response)
}

fn build_outbound_prompt(message: &str, context_player: &str, context_body: &str) -> String {
    format!(
        "Context:\n{}: {}\n\nUser:\n{}",
        context_player, context_body, message
    )
}

fn extract_text(resp: &GeminiResponse) -> Result<String, String> {
    let candidates = resp
        .candidates
        .as_ref()
        .ok_or_else(|| "gemini: no candidates in response".to_string())?;

    let candidate = candidates
        .first()
        .ok_or_else(|| "gemini: empty candidates".to_string())?;

    let content = candidate
        .content
        .as_ref()
        .ok_or_else(|| "gemini: no content in candidate".to_string())?;

    let parts = content
        .parts
        .as_ref()
        .ok_or_else(|| "gemini: no parts in content".to_string())?;

    let mut texts = Vec::new();
    for part in parts {
        if let Some(text) = &part.text {
            if !text.is_empty() {
                texts.push(text.as_str());
            }
        }
    }

    Ok(texts.join("").trim().to_string())
}
