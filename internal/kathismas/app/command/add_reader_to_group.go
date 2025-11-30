package command

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type AddReaderToGroup struct {
	GroupID    uuid.UUID
	Username   string
	TelegramID int64
	Phone      string
}

type AddReaderToGroupHandler struct {
	groupRepo domain.RepositoryReaderGroup
}

func NewAddReaderToGroupHandler(groupRepo domain.RepositoryReaderGroup) AddReaderToGroupHandler {
	if groupRepo == nil {
		panic("nil groupRepo")
	}
	return AddReaderToGroupHandler{groupRepo: groupRepo}
}

func (h AddReaderToGroupHandler) Handle(ctx context.Context, cmd AddReaderToGroup) error {
	group, err := h.groupRepo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get reader group: %w", err)
	}

	reader, err := domain.NewPsalmReader(cmd.Username, cmd.TelegramID, cmd.Phone)
	if err != nil {
		return fmt.Errorf("failed to create psalm reader: %w", err)
	}

	if err := group.AddReader(*reader); err != nil {
		return fmt.Errorf("failed to add reader to group: %w", err)
	}

	if err := h.groupRepo.Update(ctx, group); err != nil {
		return fmt.Errorf("failed to update reader group: %w", err)
	}

	return nil
}
