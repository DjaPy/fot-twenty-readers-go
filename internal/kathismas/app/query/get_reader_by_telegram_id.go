package query

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type GetReaderByTelegramIDQuery struct {
	TelegramID int64
}

type GetReaderByTelegramIDResult struct {
	GroupID      uuid.UUID
	GroupName    string
	ReaderID     uuid.UUID
	ReaderNumber int
	Username     string
}

type GetReaderByTelegramIDHandler struct {
	readerGroupRepo domain.RepositoryReaderGroup
}

func NewGetReaderByTelegramIDHandler(readerGroupRepo domain.RepositoryReaderGroup) GetReaderByTelegramIDHandler {
	return GetReaderByTelegramIDHandler{
		readerGroupRepo: readerGroupRepo,
	}
}

func (h *GetReaderByTelegramIDHandler) Handle(ctx context.Context, q *GetReaderByTelegramIDQuery) (*GetReaderByTelegramIDResult, error) {
	groups, err := h.readerGroupRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	for _, group := range groups {
		for _, reader := range group.Readers {
			if reader.TelegramID == q.TelegramID {
				return &GetReaderByTelegramIDResult{
					GroupID:      group.ID,
					GroupName:    group.Name,
					ReaderID:     reader.ID,
					ReaderNumber: int(reader.ReaderNumber),
					Username:     reader.Username,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("reader with telegram ID %d not found", q.TelegramID)
}
