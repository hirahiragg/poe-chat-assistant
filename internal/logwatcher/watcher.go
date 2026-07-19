package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

type Watcher struct {
	path     string
	pollRate time.Duration
}

func New(path string) *Watcher {
	return &Watcher{
		path:     path,
		pollRate: 250 * time.Millisecond,
	}
}

func (w *Watcher) ReadTail(maxBytes int64) ([]string, error) {
	return w.ReadRange(0, maxBytes)
}

func (w *Watcher) ReadRange(fromEnd, size int64) ([]string, error) {
	f, err := os.Open(w.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	end := info.Size() - fromEnd
	if end <= 0 {
		return nil, nil
	}
	start := end - size
	if start < 0 {
		start = 0
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	if start > 0 {
		scanner.Scan()
	}

	var lines []string
	var bytesRead int64
	for scanner.Scan() {
		bytesRead += int64(len(scanner.Bytes())) + 1
		if start+bytesRead > end {
			break
		}
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func (w *Watcher) Watch(ctx context.Context, handler func(string)) error {
	f, err := os.Open(w.path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(w.pollRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				if line != "" {
					handler(line)
				}
			}
		}
	}
}
