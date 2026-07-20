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

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Message {
    pub timestamp: NaiveDateTime,
    pub channel: Channel,
    pub guild: String,
    pub player: String,
    pub body: String,
}
