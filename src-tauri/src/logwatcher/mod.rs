use std::io::{BufRead, BufReader, Read, Seek, SeekFrom};
use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

pub struct Watcher {
    path: PathBuf,
}

impl Watcher {
    pub fn new(path: &str) -> Self {
        Self {
            path: PathBuf::from(path),
        }
    }

    /// Read the last `max_bytes` bytes of the file and return parsed lines.
    pub fn read_tail(&self, max_bytes: i64) -> std::io::Result<Vec<String>> {
        self.read_range(0, max_bytes)
    }

    /// Read a range of bytes from the file, measured from the end.
    ///
    /// `from_end`: how many bytes from the end to start (0 = at the end).
    /// `size`: how many bytes to read backwards from that point.
    ///
    /// Example: `read_range(0, 512*1024)` reads the last 512 KB.
    ///          `read_range(512*1024, 512*1024)` reads the 512 KB before that.
    pub fn read_range(&self, from_end: i64, size: i64) -> std::io::Result<Vec<String>> {
        let file = std::fs::File::open(&self.path)?;
        let metadata = file.metadata()?;
        let file_size = metadata.len() as i64;

        let end = file_size - from_end;
        if end <= 0 {
            return Ok(Vec::new());
        }

        let mut start = end - size;
        if start < 0 {
            start = 0;
        }

        let mut reader = BufReader::new(file);
        reader.seek(SeekFrom::Start(start as u64))?;

        // If we didn't start at the beginning of the file,
        // skip the first partial line.
        if start > 0 {
            let mut discard = String::new();
            reader.read_line(&mut discard)?;
        }

        let mut lines = Vec::new();
        let mut bytes_read: i64 = 0;

        loop {
            let mut line = String::new();
            let n = reader.read_line(&mut line)?;
            if n == 0 {
                break; // EOF
            }
            bytes_read += n as i64;
            if start + bytes_read > end {
                break;
            }
            let trimmed = line
                .trim_end_matches(|c: char| c == '\n' || c == '\r')
                .to_string();
            if !trimmed.is_empty() {
                lines.push(trimmed);
            }
        }

        Ok(lines)
    }

    /// Watch the file for new lines, calling `handler` for each new line.
    /// Runs in a blocking loop with 250ms polling until `cancel` is set to true.
    /// Intended to be called from `std::thread::spawn`.
    pub fn watch(
        &self,
        cancel: Arc<AtomicBool>,
        handler: impl Fn(String),
    ) -> std::io::Result<()> {
        let mut file = std::fs::File::open(&self.path)?;
        let mut pos = file.seek(SeekFrom::End(0))?;

        let poll_rate = std::time::Duration::from_millis(250);
        let mut leftover = String::new();

        while !cancel.load(Ordering::Relaxed) {
            std::thread::sleep(poll_rate);

            // Check current file size
            let metadata = file.metadata()?;
            let len = metadata.len();

            if len < pos {
                // File was truncated (e.g., log rotation); reset
                pos = len;
                leftover.clear();
                continue;
            }
            if len == pos {
                continue; // No new data
            }

            // Seek to where we left off and read new data
            file.seek(SeekFrom::Start(pos))?;
            let to_read = (len - pos) as usize;
            let mut buf = vec![0u8; to_read];
            let n = file.read(&mut buf)?;
            buf.truncate(n);
            pos += n as u64;

            // Combine with any leftover partial line
            let text = String::from_utf8_lossy(&buf);
            leftover.push_str(&text);

            // Split into complete lines
            while let Some(newline_pos) = leftover.find('\n') {
                let line = &leftover[..newline_pos];
                let trimmed = line.trim_end_matches('\r');
                if !trimmed.is_empty() {
                    handler(trimmed.to_string());
                }
                leftover = leftover[newline_pos + 1..].to_string();
            }
            // Any remaining data without '\n' stays in leftover for the next poll
        }

        Ok(())
    }
}
