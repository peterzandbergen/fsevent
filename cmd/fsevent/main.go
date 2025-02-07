package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type options struct {
	watchFile string
	logFormat string
}

var defaultOptions = options{
	watchFile: "./watchFile",
	logFormat: "TEXT",
}

func (opts *options) parseEnv(getenv func(string) string) {
	if v := getenv("FSEVENT_WATCH_FILE"); v != "" {
		opts.watchFile = v
	}
	if v := getenv("FSEVENT_LOG_FORMAT"); v != "" {
		opts.logFormat = v
	}
}

func newLogHandler(w io.Writer, format string) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // slog default
	}
	switch strings.ToUpper(format) {
	case "TEXT":
		return slog.NewTextHandler(w, opts)
	case "JSON":
		return slog.NewJSONHandler(w, opts)
	default:
		return slog.NewTextHandler(w, opts)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer, stderr io.Writer) error {
	defaultOptions.parseEnv(getenv)

	// create a logger
	logger := slog.New(newLogHandler(stderr, defaultOptions.logFormat)).With(
		slog.String("application", args[0]),
		slog.String("hostname", func() string {
			hn := "err-hostname"
			if v, err := os.Hostname(); err == nil {
				hn = v
			}
			return hn
		}()))
	slog.SetDefault(logger)

	if len(args) < 2 {
		return errors.New("missing watch path")
	}

	watcher, err := fsnotify.NewBufferedWatcher(1)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Add a watch
	path, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}

	if err := watcher.Add(path); err != nil {
		return err
	}
	logger.Info("added watch",
		slog.String("path", path),
	)

	var done = make(chan struct{}, 1)
	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Info("something's changed",
					slog.Group("event",
						slog.String("name", event.Name),
						slog.String("op", event.Op.String()),
					),
				)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Info("something's wrong",
					slog.Group("event",
						slog.String("", err.Error()),
					),
				)
			case <-ctx.Done():
				logger.Info("context done")
				return
			}
		}
	}()

	<-done
	logger.Info("received done")

	return nil
}

func main() {
	if err := run(context.Background(), os.Args, os.Getenv, os.Stdout, os.Stderr); err != nil {
		slog.Default().Info("run error",
			slog.String("error", err.Error()),
		)
	}
}
