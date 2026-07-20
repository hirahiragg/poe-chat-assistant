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
