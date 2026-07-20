use std::sync::LazyLock;

use chrono::NaiveDateTime;
use regex::Regex;

use super::model::{Channel, Message};

static LOG_LINE_RE: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(
        r"^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \d+ [0-9a-f]+ \[INFO Client \d+\] (.+)$",
    )
    .unwrap()
});

static GUILD_RE: LazyLock<Regex> = LazyLock::new(|| Regex::new(r"^<([^>]+)>\s*").unwrap());

pub fn parse_line(line: &str) -> Option<Message> {
    let caps = LOG_LINE_RE.captures(line)?;
    let ts_str = caps.get(1)?.as_str();
    let ts = NaiveDateTime::parse_from_str(ts_str, "%Y/%m/%d %H:%M:%S").ok()?;

    let payload = caps.get(2)?.as_str();

    let prefixes: &[(&str, Channel)] = &[
        ("@From ", Channel::WhisperIn),
        ("@To ", Channel::WhisperOut),
        ("#", Channel::Global),
        ("$", Channel::Trade),
        ("%", Channel::Party),
        ("&", Channel::Guild),
    ];

    let (channel, rest) = prefixes
        .iter()
        .find_map(|(prefix, channel)| {
            payload
                .strip_prefix(prefix)
                .map(|rest| (channel.clone(), rest))
        })?;

    // Parse optional guild tag: <GuildName>
    let (guild, rest) = if let Some(gcaps) = GUILD_RE.captures(rest) {
        let guild_name = gcaps.get(1)?.as_str().to_string();
        let full_match_len = gcaps.get(0)?.len();
        (guild_name, &rest[full_match_len..])
    } else {
        (String::new(), rest)
    };

    // Find "player: body" separator
    let idx = rest.find(": ")?;
    let player = rest[..idx].to_string();
    let body = rest[idx + 2..].to_string();

    Some(Message {
        timestamp: ts,
        channel,
        guild,
        player,
        body,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::NaiveDate;

    fn make_ts(y: i32, m: u32, d: u32, h: u32, min: u32, s: u32) -> NaiveDateTime {
        NaiveDate::from_ymd_opt(y, m, d)
            .unwrap()
            .and_hms_opt(h, min, s)
            .unwrap()
    }

    #[test]
    fn test_global_chat() {
        let input =
            "2025/07/12 04:26:51 46643250 cff945b9 [INFO Client 51360] #ReallyFurious: sec";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 4, 26, 51));
        assert_eq!(msg.channel, Channel::Global);
        assert_eq!(msg.guild, "");
        assert_eq!(msg.player, "ReallyFurious");
        assert_eq!(msg.body, "sec");
    }

    #[test]
    fn test_global_chat_with_guild() {
        let input = "2025/07/12 04:30:18 46850203 cff945b9 [INFO Client 51360] #<PTKFGS> Teairra_Merc: i swear, maps influence what mercs u find";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 4, 30, 18));
        assert_eq!(msg.channel, Channel::Global);
        assert_eq!(msg.guild, "PTKFGS");
        assert_eq!(msg.player, "Teairra_Merc");
        assert_eq!(
            msg.body,
            "i swear, maps influence what mercs u find"
        );
    }

    #[test]
    fn test_whisper_inbound_with_guild() {
        let input = r#"2025/07/12 04:46:18 47809953 cff945b9 [INFO Client 51360] @From <ANgRY> RuxMerc: Hi, I would like to buy your level 3 20% Enlighten Support listed for 4 divine in Mercenaries (stash tab "Trade"; position: left 11, top 12)"#;
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 4, 46, 18));
        assert_eq!(msg.channel, Channel::WhisperIn);
        assert_eq!(msg.guild, "ANgRY");
        assert_eq!(msg.player, "RuxMerc");
        assert_eq!(
            msg.body,
            r#"Hi, I would like to buy your level 3 20% Enlighten Support listed for 4 divine in Mercenaries (stash tab "Trade"; position: left 11, top 12)"#
        );
    }

    #[test]
    fn test_whisper_inbound_without_guild() {
        let input = "2025/07/12 05:44:37 51308796 cff945b9 [INFO Client 51360] @From DarkHolyhole: Hi, I would like to buy your level 3 4% Enlighten Support listed for 100 chaos";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 5, 44, 37));
        assert_eq!(msg.channel, Channel::WhisperIn);
        assert_eq!(msg.guild, "");
        assert_eq!(msg.player, "DarkHolyhole");
        assert_eq!(
            msg.body,
            "Hi, I would like to buy your level 3 4% Enlighten Support listed for 100 chaos"
        );
    }

    #[test]
    fn test_whisper_outbound() {
        let input =
            "2025/07/12 04:46:46 47838359 cff945b9 [INFO Client 51360] @To RuxMerc: ty";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 4, 46, 46));
        assert_eq!(msg.channel, Channel::WhisperOut);
        assert_eq!(msg.guild, "");
        assert_eq!(msg.player, "RuxMerc");
        assert_eq!(msg.body, "ty");
    }

    #[test]
    fn test_trade_chat() {
        let input = "2025/07/14 17:52:51 267802812 cff945b9 [INFO Client 119368] $SumoUndies: any1 could help me get trough campaign starting at act 6 ? i can pay :)";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 14, 17, 52, 51));
        assert_eq!(msg.channel, Channel::Trade);
        assert_eq!(msg.guild, "");
        assert_eq!(msg.player, "SumoUndies");
        assert_eq!(
            msg.body,
            "any1 could help me get trough campaign starting at act 6 ? i can pay :)"
        );
    }

    #[test]
    fn test_party_chat_with_guild_unicode() {
        let input = "2025/07/25 21:30:17 168597312 cff945b9 [INFO Client 59180] %<\u{00AE}\u{00C6}\u{00E5}> \u{5168}\u{8EAB}\u{53EA}\u{5269}\u{4E0B}\u{8B77}\u{7532}\u{9019}\u{96BB}\u{61C9}\u{8A72}\u{7D44}\u{7684}\u{8D77}\u{4F86}\u{5427}: TY";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 25, 21, 30, 17));
        assert_eq!(msg.channel, Channel::Party);
        assert_eq!(msg.guild, "\u{00AE}\u{00C6}\u{00E5}");
        assert_eq!(
            msg.player,
            "\u{5168}\u{8EAB}\u{53EA}\u{5269}\u{4E0B}\u{8B77}\u{7532}\u{9019}\u{96BB}\u{61C9}\u{8A72}\u{7D44}\u{7684}\u{8D77}\u{4F86}\u{5427}"
        );
        assert_eq!(msg.body, "TY");
    }

    #[test]
    fn test_japanese_global_chat() {
        let input = "2025/07/12 05:59:09 52180843 cff945b9 [INFO Client 51360] #\u{5F31}\u{6BD2}\u{6027}\u{30D7}\u{30C1}: \u{304A}\u{306F}\u{3088}\u{3046}\u{3054}\u{3058}\u{3042}\u{307E}\u{3059}";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 5, 59, 9));
        assert_eq!(msg.channel, Channel::Global);
        assert_eq!(msg.guild, "");
        assert_eq!(
            msg.player,
            "\u{5F31}\u{6BD2}\u{6027}\u{30D7}\u{30C1}"
        );
        assert_eq!(
            msg.body,
            "\u{304A}\u{306F}\u{3088}\u{3046}\u{3054}\u{3058}\u{3042}\u{307E}\u{3059}"
        );
    }

    #[test]
    fn test_guild_with_special_chars() {
        let input = "2025/07/12 23:20:23 114655390 cff945b9 [INFO Client 88228] @From <\u{00BF}ZXC?\u{2122}> Tiellonies: Hi, I would like to buy your map";
        let msg = parse_line(input).expect("should parse");
        assert_eq!(msg.timestamp, make_ts(2025, 7, 12, 23, 20, 23));
        assert_eq!(msg.channel, Channel::WhisperIn);
        assert_eq!(msg.guild, "\u{00BF}ZXC?\u{2122}");
        assert_eq!(msg.player, "Tiellonies");
        assert_eq!(msg.body, "Hi, I would like to buy your map");
    }

    #[test]
    fn test_system_message_not_chat() {
        let input = "2025/07/12 04:25:46 46578359 cff945b9 [INFO Client 51360] : You have entered Karui Shores.";
        assert!(parse_line(input).is_none());
    }

    #[test]
    fn test_engine_log_not_chat() {
        let input = "2025/07/12 04:25:03 46535421 f2498ec7 [INFO Client 51360] [JOB] Irrecoverable Exception Callback: SET";
        assert!(parse_line(input).is_none());
    }

    #[test]
    fn test_empty_line() {
        assert!(parse_line("").is_none());
    }

    #[test]
    fn test_log_file_opening() {
        let input = "2025/07/12 04:25:03 ***** LOG FILE OPENING *****";
        assert!(parse_line(input).is_none());
    }
}
