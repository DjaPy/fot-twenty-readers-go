package command

import (
	"context"
	"log/slog"

	"github.com/DjaPy/fot-twenty-readers-go/internal/common/decorator"
	"github.com/DjaPy/fot-twenty-readers-go/internal/common/errors"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
)

type CreatePsalmReaderTG struct {
	Username     string
	ReaderNumber int8
	TelegramID   int64
	Phone        string
}

type CreateCalendarOfReaderHandler decorator.CommandHandler[CreatePsalmReaderTG]

type createPsalmReaderTGHandler struct {
	repo *adapters.PsalmReaderTGRepository
}

func NewCreatePsalmReaderTGHandler(
	repo *adapters.PsalmReaderTGRepository,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) decorator.CommandHandler[CreatePsalmReaderTG] {
	return decorator.ApplyCommandDecorators[CreatePsalmReaderTG](
		createPsalmReaderTGHandler{repo: repo},
		logger,
		metricsClient,
	)
}

func (cpr createPsalmReaderTGHandler) Handle(ctx context.Context, cmd CreatePsalmReaderTG) error {

	prTG, err := domain.NewPsalmReader(cmd.Username, cmd.TelegramID, cmd.Phone, cmd.ReaderNumber)
	if err != nil {
		return err //nolint:wrapcheck // err repeat
	}

	err = cpr.repo.CreatePsalmReaderTG(ctx, prTG)
	if err != nil {
		return errors.NewSlugError(err.Error(), "unable-to-create-psalm-reader-tg-availability")
	}
	return nil
}
