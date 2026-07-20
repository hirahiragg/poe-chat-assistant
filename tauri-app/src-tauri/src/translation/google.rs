use super::TranslationRequest;

pub async fn translate(
    client: &reqwest::Client,
    req: &TranslationRequest,
) -> Result<String, String> {
    let (sl, tl) = if req.direction == super::Direction::Outbound {
        (req.target_lang.as_str(), "en")
    } else {
        ("auto", req.target_lang.as_str())
    };

    let url = reqwest::Url::parse_with_params(
        "https://translate.googleapis.com/translate_a/single",
        &[
            ("client", "gtx"),
            ("sl", sl),
            ("tl", tl),
            ("dt", "t"),
            ("q", req.message.as_str()),
        ],
    )
    .map_err(|e| format!("google translate url: {}", e))?;

    let resp = client
        .get(url)
        .header("User-Agent", "Mozilla/5.0")
        .send()
        .await
        .map_err(|e| format!("google translate request: {}", e))?;

    let status = resp.status();
    let body = resp
        .text()
        .await
        .map_err(|e| format!("google translate read: {}", e))?;

    if !status.is_success() {
        return Err(format!("google translate status {}: {}", status, body));
    }

    parse_response(&body)
}

fn parse_response(body: &str) -> Result<String, String> {
    let data: serde_json::Value =
        serde_json::from_str(body).map_err(|e| format!("google translate parse: {}", e))?;

    let sentences = data
        .as_array()
        .and_then(|arr| arr.first())
        .and_then(|v| v.as_array())
        .ok_or_else(|| "google translate: unexpected response format".to_string())?;

    let mut result = String::new();
    for sentence in sentences {
        if let Some(text) = sentence
            .as_array()
            .and_then(|parts| parts.first())
            .and_then(|v| v.as_str())
        {
            result.push_str(text);
        }
    }

    Ok(result)
}
