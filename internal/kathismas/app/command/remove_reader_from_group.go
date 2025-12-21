package command

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type RemoveReaderFromGroup struct {
	GroupID  uuid.UUID
	ReaderID uuid.UUID
}

type RemoveReaderFromGroupHandler struct {
	readerGroupRepo domain.RepositoryReaderGroup
}

func NewRemoveReaderFromGroupHandler(readerGroupRepo domain.RepositoryReaderGroup) RemoveReaderFromGroupHandler {
	if readerGroupRepo == nil {
		panic("nil readerGroupRepo")
	}
	return RemoveReaderFromGroupHandler{readerGroupRepo: readerGroupRepo}
}

func (h RemoveReaderFromGroupHandler) Handle(ctx context.Context, cmd RemoveReaderFromGroup) error {
	group, err := h.readerGroupRepo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get reader group: %w", err)
	}

	if err := group.RemoveReader(cmd.ReaderID); err != nil {
		return fmt.Errorf("failed to remove reader from group: %w", err)
	}

	if err := h.readerGroupRepo.Update(ctx, group); err != nil {
		return fmt.Errorf("failed to update reader group: %w", err)
	}

	return nil
}
