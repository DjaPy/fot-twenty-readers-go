package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/ports"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/ports/telegram"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/service"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

type options struct {
	Port          int    `short:"p" long:"port" description:"port to listen" default:"8080"`
	Conf          string `short:"f" long:"conf" env:"FM_CONF" default:"for_twenty_readers.yml" description:"config file (yml)"`
	Dbg           bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
	TelegramToken string `long:"telegram-token" env:"TELEGRAM_BOT_TOKEN" description:"Telegram bot token"`
}

var revision = "local"

func main() {
	slog.Info(fmt.Sprintf("For twenty readers %s\n", revision))
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	setupLog(opts.Dbg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := service.NewApplication(ctx)
	defer app.Close()

	if opts.TelegramToken != "" {
		logger := logrus.New()
		if opts.Dbg {
			logger.SetLevel(logrus.DebugLevel)
		}

		bot, err := telegram.NewBot(
			opts.TelegramToken,
			&app.Commands.AddReaderToGroup,
			&app.Queries.ListReaderGroups,
			&app.Queries.GetCurrentKathisma,
			app.Queries.GetReaderByTelegramID,
			logger,
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to create Telegram bot: %v", err))
		} else {
			go func() {
				slog.Info("Starting Telegram bot...")
				if err := bot.Start(ctx); err != nil && err != context.Canceled {
					slog.Error(fmt.Sprintf("Telegram bot error: %v", err))
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
		Conf:    config.Conf{},
		App:     app,
	}
	srv.Run(ctx, opts.Port)
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}
