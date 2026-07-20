use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub enum Channel {
    Global,
    Trade,
    Party,
    Guild,
    WhisperIn,
    WhisperOut,
}

impl Channel {
    pub fn symbol(&self) -> &str {
        match self {
            Channel::Global => "#",
            Channel::Trade => "$",
            Channel::Party => "%",
            Channel::Guild => "&",
            Channel::WhisperIn => "@",
            Channel::WhisperOut => "\u{2192}", // →
        }
    }

    pub fn display_name(&self) -> &str {
        match self {
            Channel::Global => "Global",
            Channel::Trade => "Trade",
            Channel::Party => "Party",
            Channel::Guild => "Guild",
            Channel::WhisperIn => "Whisper",
            Channel::WhisperOut => "Whisper(out)",
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Message {
    pub timestamp: NaiveDateTime,
    pub channel: Channel,
    pub guild: String,
    pub player: String,
    pub body: String,
}
