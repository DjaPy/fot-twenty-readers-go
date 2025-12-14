package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const numWorkers = 10

type Bot struct {
	api      *tgbotapi.BotAPI
	handlers *Handlers
	log      *slog.Logger
	wg       sync.WaitGroup
}

func NewBot(
	token string,
	addReaderHandler *command.AddReaderToGroupHandler,
	listGroupsHandler *query.ListReaderGroupsHandler,
	getCurrentKathismaHandler *query.GetCurrentKathismaHandler,
	getReaderByTelegramIDHandler query.GetReaderByTelegramIDHandler,
	log *slog.Logger,
) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	sessionManager := NewSessionManager()
	handlers := NewHandlers(
		sessionManager,
		addReaderHandler,
		listGroupsHandler,
		getCurrentKathismaHandler,
		getReaderByTelegramIDHandler,
		log,
	)

	log.Info("Authorized on account", "username", bot.Self.UserName)

	return &Bot{
		api:      bot,
		handlers: handlers,
		log:      log,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.log.Info("Starting Telegram bot...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	jobs := make(chan tgbotapi.Update, numWorkers)

	for i := 0; i < numWorkers; i++ {
		b.wg.Add(1)
		go b.worker(ctx, jobs)
	}

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Stopping Telegram bot...")
			close(jobs)
			b.wg.Wait()
			b.api.StopReceivingUpdates()
			return fmt.Errorf("telegram bot context finished: %w", ctx.Err())
		case update := <-updates:
			jobs <- update
		}
	}
}

func (b *Bot) worker(ctx context.Context, jobs <-chan tgbotapi.Update) {
	defer b.wg.Done()
	for update := range jobs {
		if update.Message != nil {
			b.handleUpdate(ctx, update)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message.IsCommand() {
		if err := b.handlers.HandleCommand(ctx, b.api, update.Message); err != nil {
			b.log.Error("error handling command", "error", err)
		}
	} else {
		if err := b.handlers.HandleMessage(ctx, b.api, update.Message); err != nil {
			b.log.Error("error handling message", "error", err)
		}
	}
}

func (b *Bot) handleCallbackQuery(ctx context.Context, update tgbotapi.Update) {
	if err := b.handlers.HandleCallbackQuery(ctx, b.api, update.CallbackQuery); err != nil {
		b.log.Error("error handling callback query", "error", err)
	}
}
