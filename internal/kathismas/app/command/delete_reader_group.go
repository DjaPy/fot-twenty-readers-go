package command

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type DeleteReaderGroup struct {
	GroupID uuid.UUID
}

type DeleteReaderGroupHandler struct {
	readerGroupRepo domain.RepositoryReaderGroup
}

func NewDeleteReaderGroupHandler(readerGroupRepo domain.RepositoryReaderGroup) DeleteReaderGroupHandler {
	if readerGroupRepo == nil {
		panic("nil readerGroupRepo")
	}
	return DeleteReaderGroupHandler{readerGroupRepo: readerGroupRepo}
}

func (h DeleteReaderGroupHandler) Handle(ctx context.Context, cmd DeleteReaderGroup) error {
	if _, err := h.readerGroupRepo.GetByID(ctx, cmd.GroupID); err != nil {
		return fmt.Errorf("failed to get reader group: %w", err)
	}

	if err := h.readerGroupRepo.Delete(ctx, cmd.GroupID); err != nil {
		return fmt.Errorf("failed to delete reader group: %w", err)
	}

	return nil
}
