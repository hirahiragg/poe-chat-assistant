package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/daichirata/poe-chat-assistant/internal/chat"
	"github.com/daichirata/poe-chat-assistant/internal/logwatcher"
)

func main() {
	path := flag.String("log", "", "path to Client.txt")
	flag.Parse()

	if *path == "" {
		fmt.Fprintln(os.Stderr, "Usage: poe-chat-assistant -log /path/to/Client.txt")
		os.Exit(1)
	}

	store := chat.NewStore(50)
	watcher := logwatcher.New(*path)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Watching %s ...\n", *path)

	err := watcher.Watch(ctx, func(line string) {
		msg, ok := chat.ParseLine(line)
		if !ok {
			return
		}
		store.Add(msg)
		fmt.Println(msg)
	})

	if err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
