package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"gioui.org/app"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
	"github.com/hirahiragg/poe-chat-assistant/internal/config"
	"github.com/hirahiragg/poe-chat-assistant/internal/logwatcher"
	"github.com/hirahiragg/poe-chat-assistant/internal/translation"
	"github.com/hirahiragg/poe-chat-assistant/ui"
)

func main() {
	mode := flag.String("mode", "ui", "mode: ui, watch, translate")
	logPath := flag.String("log", "", "path to Client.txt")
	translator := flag.String("translator", "", "translator: google, deepl, gemini")
	apiKey := flag.String("api-key", "", "API key for deepl or gemini")
	flag.Parse()

	switch *mode {
	case "ui":
		go runUI(*logPath, *translator, *apiKey)
		app.Main()
		return
	case "watch":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		runWatch(ctx, *logPath)
	case "translate":
		cfg := config.Load()
		if *translator != "" {
			cfg.Translator = *translator
		}
		key := *apiKey
		if key == "" {
			key = cfg.APIKey()
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		runTranslate(ctx, cfg.Translator, key)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

func newTranslator(name, apiKey string) translation.Translator {
	switch name {
	case "deepl":
		if apiKey == "" {
			apiKey = os.Getenv("DEEPL_API_KEY")
		}
		if apiKey == "" {
			log.Printf("warning: DeepL selected but no API key configured")
			return translation.NewGoogle()
		}
		return translation.NewDeepL(apiKey)
	case "gemini":
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			log.Printf("warning: Gemini selected but no API key configured")
			return translation.NewGoogle()
		}
		t, err := translation.NewGemini(context.Background(), apiKey)
		if err != nil {
			log.Printf("warning: gemini init failed: %v, falling back to google", err)
			return translation.NewGoogle()
		}
		return t
	default:
		return translation.NewGoogle()
	}
}

func runUI(flagLogPath, flagTranslator, flagAPIKey string) {
	cfg := config.Load()

	if flagLogPath != "" {
		cfg.LogPath = flagLogPath
	}
	if flagTranslator != "" {
		cfg.Translator = flagTranslator
	}
	if flagAPIKey != "" {
		switch cfg.Translator {
		case "deepl":
			cfg.DeepLKey = flagAPIKey
		case "gemini":
			cfg.GeminiKey = flagAPIKey
		}
	}

	store := chat.NewStore(50)

	var mu sync.Mutex
	svc := translation.NewService(newTranslator(cfg.Translator, cfg.APIKey()))

	translateFn := func(msg chat.Message, replyText string, done func(inbound, outbound string)) {
		mu.Lock()
		currentSvc := svc
		mu.Unlock()

		var inbound, outbound string
		ctx := context.Background()

		if replyText == "" {
			result, err := currentSvc.Translate(ctx, translation.Request{
				Direction:  translation.Inbound,
				Message:    msg.Body,
				TargetLang: cfg.TargetLang,
			})
			if err != nil {
				inbound = fmt.Sprintf("error: %v", err)
			} else {
				inbound = result
			}
		}

		if replyText != "" {
			result, err := currentSvc.Translate(ctx, translation.Request{
				Direction:  translation.Outbound,
				Message:    replyText,
				Context:    []chat.Message{msg},
				TargetLang: cfg.TargetLang,
			})
			if err != nil {
				outbound = fmt.Sprintf("error: %v", err)
			} else {
				outbound = result
			}
		}

		done(inbound, outbound)
	}

	const chunkSize int64 = 512 * 1024
	var loadedBytes int64 = chunkSize

	loadMore := func() ui.LoadMoreResult {
		if cfg.LogPath == "" {
			return ui.LoadMoreEOF
		}
		watcher := logwatcher.New(cfg.LogPath)
		const maxAttempts = 20
		for range maxAttempts {
			lines, err := watcher.ReadRange(loadedBytes, chunkSize)
			if err != nil || len(lines) == 0 {
				return ui.LoadMoreEOF
			}
			var msgs []chat.Message
			for _, line := range lines {
				if msg, ok := chat.ParseLine(line); ok {
					msgs = append(msgs, msg)
				}
			}
			loadedBytes += chunkSize
			if len(msgs) > 0 {
				store.Prepend(msgs)
				return ui.LoadMoreFound
			}
		}
		return ui.LoadMoreNotFound
	}

	var watchCancel context.CancelFunc

	startWatcher := func(path string, application *ui.App) {
		if watchCancel != nil {
			watchCancel()
		}
		if path == "" {
			return
		}

		watcher := logwatcher.New(path)
		lines, err := watcher.ReadTail(512 * 1024)
		if err == nil {
			for _, line := range lines {
				if msg, ok := chat.ParseLine(line); ok {
					store.Add(msg)
				}
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		watchCancel = cancel
		go func() {
			w := logwatcher.New(path)
			_ = w.Watch(ctx, func(line string) {
				msg, ok := chat.ParseLine(line)
				if !ok {
					return
				}
				store.Add(msg)
				application.Window().Invalidate()
			})
		}()
	}

	onConfigSave := func(newCfg *config.Config) {
		if err := newCfg.Save(); err != nil {
			log.Printf("config save error: %v", err)
		}

		mu.Lock()
		svc = translation.NewService(newTranslator(newCfg.Translator, newCfg.APIKey()))
		mu.Unlock()

		cfg = newCfg
	}

	application := ui.NewApp(store, cfg, translateFn, onConfigSave, loadMore)

	startWatcher(cfg.LogPath, application)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}

func runWatch(ctx context.Context, path string) {
	if path == "" {
		fmt.Fprintln(os.Stderr, "Usage: poe-chat-assistant -mode watch -log /path/to/Client.txt")
		os.Exit(1)
	}

	store := chat.NewStore(50)
	watcher := logwatcher.New(path)

	fmt.Fprintf(os.Stderr, "Watching %s ...\n", path)

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

func runTranslate(ctx context.Context, translatorName, apiKey string) {
	svc := translation.NewService(newTranslator(translatorName, apiKey))
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Translation test mode")
	fmt.Println("  Type English text  -> translated to Japanese (inbound)")
	fmt.Println("  Prefix with > for outbound: >Japanese text -> translated to English")
	fmt.Println("  Ctrl+C to exit")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		dir := translation.Inbound
		msg := line
		if strings.HasPrefix(line, ">") {
			dir = translation.Outbound
			msg = strings.TrimSpace(line[1:])
		}

		result, err := svc.Translate(ctx, translation.Request{
			Direction:  dir,
			Message:    msg,
			TargetLang: "ja",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			continue
		}

		label := "-> JA"
		if dir == translation.Outbound {
			label = "-> EN"
		}
		fmt.Printf("  %s: %s\n", label, result)
	}
}
