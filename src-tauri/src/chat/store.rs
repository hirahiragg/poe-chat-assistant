use std::collections::VecDeque;

use super::model::Message;

pub struct Store {
    messages: VecDeque<Message>,
    limit: usize,
}

impl Store {
    pub fn new(limit: usize) -> Self {
        Self {
            messages: VecDeque::with_capacity(limit),
            limit,
        }
    }

    pub fn add(&mut self, msg: Message) {
        self.messages.push_back(msg);
        while self.messages.len() > self.limit {
            self.messages.pop_front();
        }
    }

    /// Prepend messages to the front of the store.
    /// `msgs` should be in chronological order (oldest first).
    pub fn prepend(&mut self, msgs: Vec<Message>) {
        for msg in msgs.into_iter().rev() {
            self.messages.push_front(msg);
        }
    }

    /// Return all messages, newest first.
    pub fn list(&self) -> Vec<Message> {
        self.messages.iter().rev().cloned().collect()
    }

    #[allow(dead_code)]
    pub fn len(&self) -> usize {
        self.messages.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::chat::model::{Channel, Message};
    use chrono::NaiveDate;

    fn make_msg(body: &str) -> Message {
        Message {
            timestamp: NaiveDate::from_ymd_opt(2025, 1, 1)
                .unwrap()
                .and_hms_opt(0, 0, 0)
                .unwrap(),
            channel: Channel::Global,
            guild: String::new(),
            player: "Test".to_string(),
            body: body.to_string(),
        }
    }

    #[test]
    fn add_and_list() {
        let mut store = Store::new(10);
        store.add(make_msg("first"));
        store.add(make_msg("second"));
        let list = store.list();
        assert_eq!(list.len(), 2);
        assert_eq!(list[0].body, "second");
        assert_eq!(list[1].body, "first");
    }

    #[test]
    fn evicts_oldest_when_over_limit() {
        let mut store = Store::new(2);
        store.add(make_msg("a"));
        store.add(make_msg("b"));
        store.add(make_msg("c"));
        assert_eq!(store.len(), 2);
        let list = store.list();
        assert_eq!(list[0].body, "c");
        assert_eq!(list[1].body, "b");
    }

    #[test]
    fn prepend_adds_to_front() {
        let mut store = Store::new(10);
        store.add(make_msg("latest"));
        store.prepend(vec![make_msg("old1"), make_msg("old2")]);
        let list = store.list();
        assert_eq!(list.len(), 3);
        assert_eq!(list[0].body, "latest");
        assert_eq!(list[1].body, "old2");
        assert_eq!(list[2].body, "old1");
    }

    #[test]
    fn list_returns_newest_first() {
        let mut store = Store::new(10);
        for i in 0..5 {
            store.add(make_msg(&format!("msg{}", i)));
        }
        let list = store.list();
        assert_eq!(list[0].body, "msg4");
        assert_eq!(list[4].body, "msg0");
    }

    #[test]
    fn empty_store() {
        let store = Store::new(10);
        assert_eq!(store.len(), 0);
        assert!(store.list().is_empty());
    }

    #[test]
    fn limit_of_one() {
        let mut store = Store::new(1);
        store.add(make_msg("a"));
        store.add(make_msg("b"));
        assert_eq!(store.len(), 1);
        assert_eq!(store.list()[0].body, "b");
    }
}
