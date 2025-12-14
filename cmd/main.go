package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/ports"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/ports/telegram"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/service"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Port          int    `short:"p" long:"port" description:"port to listen" default:"8080"`
	Conf          string `short:"f" long:"conf" env:"FM_CONF" default:"for_twenty_readers.yml" description:"config file (yml)"`
	Dbg           bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
	TelegramToken string `long:"telegram-token" env:"TELEGRAM_BOT_TOKEN" description:"Telegram bot token"`
}

var revision = "local"

func main() {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		slog.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}
	setupLog(opts.Dbg)
	logger := slog.Default()

	slog.Info("For twenty readers", "revision", revision)

	cfg, err := config.NewConfiguration()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := service.NewApplication(ctx, logger)
	defer app.Close()

	if opts.TelegramToken != "" {

		bot, err := telegram.NewBot(
			opts.TelegramToken,
			cfg.Telegram.NumWorkers,
			&app.Commands.AddReaderToGroup,
			&app.Queries.ListReaderGroups,
			&app.Queries.GetReaderGroup,
			&app.Queries.GetCurrentKathisma,
			app.Queries.GetReaderByTelegramID,
			logger,
		)
		if err != nil {
			slog.Error("failed to create Telegram bot", "error", err)
		} else {
			go func() {
				slog.Info("Starting Telegram bot...")
				if err := bot.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
					slog.Error("telegram bot error", "error", err)
				}
			}()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down gracefully...")
		cancel()
	}()

	srv := &ports.Server{
		Version: revision,
		Conf:    *cfg,
		App:     app,
	}
	srv.Run(ctx, opts.Port)
}

func setupLog(dbg bool) {
	logLevel := slog.LevelInfo
	if dbg {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))
}
