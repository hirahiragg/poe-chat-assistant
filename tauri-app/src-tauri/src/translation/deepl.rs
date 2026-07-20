use serde::Deserialize;

use super::TranslationRequest;

#[derive(Deserialize)]
struct DeepLResponse {
    translations: Vec<DeepLTranslation>,
}

#[derive(Deserialize)]
struct DeepLTranslation {
    text: String,
}

pub async fn translate(
    client: &reqwest::Client,
    api_key: &str,
    req: &TranslationRequest,
) -> Result<String, String> {
    let api_url = if api_key.contains(":fx") {
        "https://api-free.deepl.com/v2/translate"
    } else {
        "https://api.deepl.com/v2/translate"
    };

    let target_lang = if req.direction == super::Direction::Outbound {
        "EN".to_string()
    } else {
        req.target_lang.to_uppercase()
    };

    let resp = client
        .post(api_url)
        .header("Authorization", format!("DeepL-Auth-Key {}", api_key))
        .form(&[
            ("text", req.message.as_str()),
            ("target_lang", target_lang.as_str()),
        ])
        .send()
        .await
        .map_err(|e| format!("deepl request: {}", e))?;

    let status = resp.status();
    let body = resp
        .text()
        .await
        .map_err(|e| format!("deepl read: {}", e))?;

    if !status.is_success() {
        return Err(format!("deepl status {}: {}", status, body));
    }

    let result: DeepLResponse =
        serde_json::from_str(&body).map_err(|e| format!("deepl parse: {}", e))?;

    result
        .translations
        .into_iter()
        .next()
        .map(|t| t.text)
        .ok_or_else(|| "deepl: empty response".to_string())
}
