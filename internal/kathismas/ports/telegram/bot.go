package telegram

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	handlers *Handlers
	log      *logrus.Logger
}

func NewBot(
	token string,
	addReaderHandler *command.AddReaderToGroupHandler,
	listGroupsHandler *query.ListReaderGroupsHandler,
	getCurrentKathismaHandler *query.GetCurrentKathismaHandler,
	getReaderByTelegramIDHandler query.GetReaderByTelegramIDHandler,
	log *logrus.Logger,
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

	log.Infof("Authorized on account %s", bot.Self.UserName)

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

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Stopping Telegram bot...")
			b.api.StopReceivingUpdates()
			return fmt.Errorf("telegram bot context finished: %w", ctx.Err())
		case update := <-updates:
			if update.Message != nil {
				go b.handleUpdate(ctx, update)
			} else if update.CallbackQuery != nil {
				go b.handleCallbackQuery(ctx, update)
			}
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message.IsCommand() {
		if err := b.handlers.HandleCommand(ctx, b.api, update.Message); err != nil {
			b.log.Errorf("Error handling command: %v", err)
		}
	} else {
		if err := b.handlers.HandleMessage(ctx, b.api, update.Message); err != nil {
			b.log.Errorf("Error handling message: %v", err)
		}
	}
}

func (b *Bot) handleCallbackQuery(ctx context.Context, update tgbotapi.Update) {
	if err := b.handlers.HandleCallbackQuery(ctx, b.api, update.CallbackQuery); err != nil {
		b.log.Errorf("Error handling callback query: %v", err)
	}
}
