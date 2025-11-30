package command

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type CreateReaderGroup struct {
	Name        string
	StartOffset int
}

type CreateReaderGroupHandler struct {
	repo domain.RepositoryReaderGroup
}

func NewCreateReaderGroupHandler(repo domain.RepositoryReaderGroup) CreateReaderGroupHandler {
	if repo == nil {
		panic("nil repo")
	}
	return CreateReaderGroupHandler{repo: repo}
}

func (h CreateReaderGroupHandler) Handle(ctx context.Context, cmd CreateReaderGroup) (uuid.UUID, error) {
	group, err := domain.NewReaderGroup(cmd.Name, cmd.StartOffset)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create reader group: %w", err)
	}

	if err := h.repo.Create(ctx, group); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save reader group: %w", err)
	}

	return group.ID, nil
}
