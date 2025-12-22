package command

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type UpdateReaderGroup struct {
	GroupID     uuid.UUID
	Name        *string
	StartOffset *int
}

type UpdateReaderGroupHandler struct {
	readerGroupRepo domain.RepositoryReaderGroup
}

func NewUpdateReaderGroupHandler(readerGroupRepo domain.RepositoryReaderGroup) UpdateReaderGroupHandler {
	if readerGroupRepo == nil {
		panic("nil readerGroupRepo")
	}
	return UpdateReaderGroupHandler{readerGroupRepo: readerGroupRepo}
}

func (h UpdateReaderGroupHandler) Handle(ctx context.Context, cmd UpdateReaderGroup) error {
	group, err := h.readerGroupRepo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get reader group: %w", err)
	}

	if cmd.Name != nil && cmd.StartOffset != nil {
		if err := group.UpdateName(*cmd.Name); err != nil {
			return fmt.Errorf("failed to update group name: %w", err)
		}
		if err := group.UpdateStartOffset(*cmd.StartOffset); err != nil {
			return fmt.Errorf("failed to update start offset: %w", err)
		}
	} else if cmd.Name != nil {
		if err := group.UpdateName(*cmd.Name); err != nil {
			return fmt.Errorf("failed to update group name: %w", err)
		}
	} else if cmd.StartOffset != nil {
		if err := group.UpdateStartOffset(*cmd.StartOffset); err != nil {
			return fmt.Errorf("failed to update start offset: %w", err)
		}
	}

	errUpd := h.readerGroupRepo.Update(ctx, group)
	if errUpd != nil {
		return fmt.Errorf("failed to update reader group: %w", errUpd)
	}
	return nil
}
