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
