package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofrs/uuid/v5"
)

type MessageSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

type Handlers struct {
	sessionManager               *SessionManager
	addReaderHandler             *command.AddReaderToGroupHandler
	listGroupsHandler            *query.ListReaderGroupsHandler
	getReaderGroupHandler        *query.GetReaderGroupHandler
	getCurrentKathismaHandler    *query.GetCurrentKathismaHandler
	getReaderByTelegramIDHandler query.GetReaderByTelegramIDHandler
	log                          *slog.Logger
}

func NewHandlers(
	sessionManager *SessionManager,
	addReaderHandler *command.AddReaderToGroupHandler,
	listGroupsHandler *query.ListReaderGroupsHandler,
	getReaderGroupHandler *query.GetReaderGroupHandler,
	getCurrentKathismaHandler *query.GetCurrentKathismaHandler,
	getReaderByTelegramIDHandler query.GetReaderByTelegramIDHandler,
	log *slog.Logger,
) *Handlers {
	return &Handlers{
		sessionManager:               sessionManager,
		addReaderHandler:             addReaderHandler,
		listGroupsHandler:            listGroupsHandler,
		getReaderGroupHandler:        getReaderGroupHandler,
		getCurrentKathismaHandler:    getCurrentKathismaHandler,
		getReaderByTelegramIDHandler: getReaderByTelegramIDHandler,
		log:                          log,
	}
}

func (h *Handlers) HandleCommand(ctx context.Context, bot MessageSender, message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		return h.handleStart(bot, message)
	case "register":
		return h.handleRegister(ctx, bot, message)
	case "kathisma":
		return h.handleKathisma(ctx, bot, message)
	case "cancel":
		return h.handleCancel(bot, message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используйте /start для начала.")
		_, err := bot.Send(msg)
		return fmt.Errorf("failed to send unknown command message: %w", err)
	}
}

func (h *Handlers) handleStart(bot MessageSender, message *tgbotapi.Message) error {
	welcomeText := `Добро пожаловать в бот для чтецов Псалтири!

Доступные команды:
/register - Зарегистрироваться в группе
/kathisma - Узнать текущую кафизму
/cancel - Отменить регистрацию`

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	_, err := bot.Send(msg)
	return fmt.Errorf("failed to send start message: %w", err)
}

func (h *Handlers) handleRegister(ctx context.Context, bot MessageSender, message *tgbotapi.Message) error {
	session := h.sessionManager.GetSession(message.From.ID)

	if session.State != StateIdle {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Регистрация уже в процессе. Используйте /cancel для отмены.")
		_, err := bot.Send(msg)
		return fmt.Errorf("registration already in progress: %w", err)
	}

	readerInfo, err := h.getReaderByTelegramIDHandler.Handle(ctx, &query.GetReaderByTelegramIDQuery{
		TelegramID: message.From.ID,
	})

	if err == nil {
		responseText := fmt.Sprintf("Вы уже зарегистрированы!\n\nГруппа: %s\nВаш номер: %d\n\nИспользуйте /kathisma для просмотра текущей кафизмы.",
			readerInfo.GroupName, readerInfo.ReaderNumber)
		msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
		_, sendErr := bot.Send(msg)
		return fmt.Errorf("failed to send registration confirmation: %w", sendErr)
	}

	h.sessionManager.UpdateState(message.From.ID, StateAwaitingName)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, введите ваше имя:")
	_, err = bot.Send(msg)
	return fmt.Errorf("failed to send name input prompt: %w", err)
}

func (h *Handlers) handleKathisma(ctx context.Context, bot MessageSender, message *tgbotapi.Message) error {
	readerInfo, err := h.getReaderByTelegramIDHandler.Handle(ctx, &query.GetReaderByTelegramIDQuery{
		TelegramID: message.From.ID,
	})

	if err != nil {
		h.log.Info("reader not found for telegram ID", "telegram_id", message.From.ID, "error", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы не зарегистрированы. Используйте /register для регистрации.")
		_, sendErr := bot.Send(msg)
		return fmt.Errorf("failed to send message: %w", sendErr)
	}

	return h.handleGetKathismaForRegistered(ctx, bot, message, readerInfo.GroupID, readerInfo.ReaderNumber)
}

func (h *Handlers) handleCancel(bot MessageSender, message *tgbotapi.Message) error {
	h.sessionManager.DeleteSession(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Регистрация отменена.")
	_, err := bot.Send(msg)
	return fmt.Errorf("failed to send cancel message: %w", err)
}

func (h *Handlers) HandleMessage(ctx context.Context, bot MessageSender, message *tgbotapi.Message) error {
	session := h.sessionManager.GetSession(message.From.ID)

	switch session.State {
	case StateAwaitingName:
		return h.handleNameInput(ctx, bot, message)
	case StateAwaitingGroup:
		return h.handleGroupSelection(bot, message)
	case StateAwaitingConfirm:
		return h.handleConfirmation(bot, message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Используйте /start для начала работы с ботом.")
		_, err := bot.Send(msg)
		return fmt.Errorf("failed to send default message: %w", err)
	}
}

func (h *Handlers) HandleCallbackQuery(ctx context.Context, bot MessageSender, callback *tgbotapi.CallbackQuery) error {
	session := h.sessionManager.GetSession(callback.From.ID)

	switch session.State {
	case StateAwaitingGroup:
		return h.handleGroupCallback(ctx, bot, callback)
	case StateAwaitingReaderNumber:
		return h.handleReaderNumberCallback(bot, callback)
	case StateAwaitingConfirm:
		return h.handleConfirmCallback(ctx, bot, callback)
	default:
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неожиданный callback")
		_, err := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", err)
	}
}

func (h *Handlers) handleNameInput(ctx context.Context, bot MessageSender, message *tgbotapi.Message) error {
	name := strings.TrimSpace(message.Text)
	if name == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Имя не может быть пустым. Пожалуйста, введите ваше имя:")
		_, err := bot.Send(msg)
		return fmt.Errorf("failed to send empty name message: %w", err)
	}

	session := h.sessionManager.GetSession(message.From.ID)
	session.Username = name
	session.State = StateAwaitingGroup
	h.sessionManager.SetSession(message.From.ID, session)

	groups, err := h.listGroupsHandler.Handle(ctx, query.ListReaderGroups{})
	if err != nil {
		h.log.Error("failed to list groups", "error", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при загрузке списка групп. Попробуйте позже.")
		_, err = bot.Send(msg)
		return fmt.Errorf("failed to send error message after listing groups: %w", err)
	}

	if len(groups) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "В системе пока нет групп. Обратитесь к администратору.")
		h.sessionManager.DeleteSession(message.From.ID)
		_, err = bot.Send(msg)
		return fmt.Errorf("failed to send no groups message: %w", err)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, group := range groups {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%d чтецов)", group.Name, group.ReadersCount),
				fmt.Sprintf("group:%s", group.ID),
			),
		)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите группу:")
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	return fmt.Errorf("failed to send group selection message: %w", err)
}

func (h *Handlers) handleGroupCallback(ctx context.Context, bot MessageSender, callback *tgbotapi.CallbackQuery) error {
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 2 || parts[0] != "group" {
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неверный формат данных")
		_, err := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", err)
	}

	groupID, err := uuid.FromString(parts[1])
	if err != nil {
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неверный ID группы")
		_, sendErr := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", sendErr)
	}

	group, err := h.getReaderGroupHandler.Handle(ctx, query.GetReaderGroup{ID: groupID})
	if err != nil {
		h.log.Error("failed to get group", "error", err)
		answerCallback := tgbotapi.NewCallback(callback.ID, "Ошибка при получении группы")
		_, sendErr := bot.Request(answerCallback)
		if sendErr != nil {
			h.log.Error("failed to answer callback", "error", sendErr)
		}

		errorMsg := fmt.Sprintf("Ошибка при получении группы: %v", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, errorMsg)
		_, sendErr = bot.Send(msg)
		return fmt.Errorf("failed to send error message: %w", sendErr)
	}

	availableNumbers := group.GetAvailableReaderNumbers()
	if len(availableNumbers) == 0 {
		h.log.Info("group is full, cannot add reader", "group_id", groupID)
		answerCallback := tgbotapi.NewCallback(callback.ID, "Группа полностью заполнена")
		_, sendErr := bot.Request(answerCallback)
		if sendErr != nil {
			h.log.Error("failed to answer callback", "error", sendErr)
		}

		errorMsg := "Группа полностью заполнена (20 чтецов). Обратитесь к администратору."
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, errorMsg)
		h.sessionManager.DeleteSession(callback.From.ID)
		_, sendErr = bot.Send(msg)
		return fmt.Errorf("failed to send full group message: %w", sendErr)
	}

	session := h.sessionManager.GetSession(callback.From.ID)
	session.GroupID = groupID
	session.GroupName = group.Name
	session.State = StateAwaitingReaderNumber
	h.sessionManager.SetSession(callback.From.ID, session)

	answerCallback := tgbotapi.NewCallback(callback.ID, "Группа выбрана")
	_, err = bot.Request(answerCallback)
	if err != nil {
		h.log.Error("failed to answer callback", "error", err)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	row := make([]tgbotapi.InlineKeyboardButton, 0, len(availableNumbers))
	for i, num := range availableNumbers {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("№%d", num),
			fmt.Sprintf("reader:%d", num),
		)
		row = append(row, btn)

		if (i+1)%4 == 0 || i == len(availableNumbers)-1 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("Группа: %s\n\nВыберите ваш номер чтеца:", group.Name))
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	return fmt.Errorf("failed to send reader number selection: %w", err)
}

func (h *Handlers) handleReaderNumberCallback(bot MessageSender, callback *tgbotapi.CallbackQuery) error {
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 2 || parts[0] != "reader" {
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неверный формат данных")
		_, err := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", err)
	}

	var readerNumber int8
	_, err := fmt.Sscanf(parts[1], "%d", &readerNumber)
	if err != nil || readerNumber < 1 || readerNumber > 20 {
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неверный номер чтеца")
		_, sendErr := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", sendErr)
	}

	session := h.sessionManager.GetSession(callback.From.ID)
	session.ReaderNumber = readerNumber
	session.State = StateAwaitingConfirm
	h.sessionManager.SetSession(callback.From.ID, session)

	answerCallback := tgbotapi.NewCallback(callback.ID, "Номер выбран")
	_, err = bot.Request(answerCallback)
	if err != nil {
		h.log.Error("failed to answer callback", "error", err)
	}

	confirmText := fmt.Sprintf("Подтвердите регистрацию:\n\nИмя: %s\nГруппа: %s\nНомер чтеца: %d\n\nВсё верно?",
		session.Username, session.GroupName, session.ReaderNumber)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить", "confirm:yes"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить", "confirm:no"),
		),
	)

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmText)
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	return fmt.Errorf("failed to send confirmation prompt: %w", err)
}

func (h *Handlers) handleGroupSelection(bot MessageSender, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, выберите группу из списка выше.")
	_, err := bot.Send(msg)
	return fmt.Errorf("failed to send group selection prompt: %w", err)
}

func (h *Handlers) handleConfirmation(bot MessageSender, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, используйте кнопки для подтверждения.")
	_, err := bot.Send(msg)
	return fmt.Errorf("failed to send confirmation prompt: %w", err)
}

func (h *Handlers) handleConfirmCallback(ctx context.Context, bot MessageSender, callback *tgbotapi.CallbackQuery) error {
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 2 || parts[0] != "confirm" {
		answerCallback := tgbotapi.NewCallback(callback.ID, "Неверный формат данных")
		_, err := bot.Request(answerCallback)
		return fmt.Errorf("failed to send callback answer: %w", err)
	}

	session := h.sessionManager.GetSession(callback.From.ID)

	if parts[1] != "yes" {
		h.sessionManager.DeleteSession(callback.From.ID)
		answerCallback := tgbotapi.NewCallback(callback.ID, "Регистрация отменена")
		_, err := bot.Request(answerCallback)
		if err != nil {
			h.log.Error("failed to answer callback", "error", err)
		}

		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Регистрация отменена. Используйте /register для повторной попытки.")
		_, err = bot.Send(msg)
		return fmt.Errorf("failed to send cancellation message after user canceled: %w", err)
	}

	cmd := command.AddReaderToGroup{
		GroupID:      session.GroupID,
		ReaderNumber: session.ReaderNumber,
		Username:     session.Username,
		TelegramID:   callback.From.ID,
		Phone:        "",
	}

	err := h.addReaderHandler.Handle(ctx, cmd)
	if err != nil {
		h.log.Error("failed to add reader", "error", err)
		answerCallback := tgbotapi.NewCallback(callback.ID, "Ошибка при регистрации")
		_, sendErr := bot.Request(answerCallback)
		if sendErr != nil {
			h.log.Error("failed to answer callback", "error", sendErr)
		}

		errorMsg := fmt.Sprintf("Ошибка при регистрации: %v\n\nПопробуйте снова через /register", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, errorMsg)
		h.sessionManager.DeleteSession(callback.From.ID)
		_, sendErr = bot.Send(msg)
		return fmt.Errorf("failed to send error message after registration failure: %w", sendErr)
	}

	h.sessionManager.DeleteSession(callback.From.ID)

	answerCallback := tgbotapi.NewCallback(callback.ID, "Регистрация успешна!")
	_, err = bot.Request(answerCallback)
	if err != nil {
		h.log.Error("failed to answer callback", "error", err)
	}

	successMsg := fmt.Sprintf("✅ Регистрация успешна!\n\nВы зарегистрированы как чтец №%d в группе \"%q\".\n\nИспользуйте /kathisma для просмотра текущей кафизмы.",
		session.ReaderNumber, session.GroupName)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, successMsg)
	_, sendErr := bot.Send(msg)
	return fmt.Errorf("failed to send success message: %w", sendErr)
}

func (h *Handlers) handleGetKathismaForRegistered(
	ctx context.Context,
	bot MessageSender,
	message *tgbotapi.Message,
	groupID uuid.UUID,
	readerNumber int,
) error {
	result, err := h.getCurrentKathismaHandler.Handle(ctx, query.GetCurrentKathisma{
		GroupID:      groupID,
		ReaderNumber: readerNumber,
	})

	if err != nil {
		h.log.Error("failed to get current kathisma", "error", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Ошибка при получении кафизмы: %v", err))
		_, sendErr := bot.Send(msg)
		return fmt.Errorf("failed to send kathisma error message: %w", sendErr)
	}

	var responseText string
	if result.Kathisma == 0 {
		responseText = fmt.Sprintf("📖 На сегодня (%s) чтение не предусмотрено.\n\n", result.Date)
	} else {
		responseText = fmt.Sprintf(
			"📖 Ваша кафизма на сегодня (%s):\n\n Кафизма №%d\n\nЧтец №%d в группе \"%q\"",
			result.Date, result.Kathisma, result.ReaderNumber, result.GroupName,
		)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	_, err = bot.Send(msg)
	return fmt.Errorf("failed to send kathisma message: %w", err)
}
