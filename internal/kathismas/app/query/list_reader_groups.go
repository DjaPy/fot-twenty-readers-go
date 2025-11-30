package query

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
)

type ListReaderGroups struct{}

type ReaderGroupDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StartOffset    int    `json:"start_offset"`
	ReadersCount   int    `json:"readers_count"`
	CalendarsCount int    `json:"calendars_count"`
	CreatedAt      string `json:"created_at"`
}

type ListReaderGroupsHandler struct {
	repo domain.RepositoryReaderGroup
}

func NewListReaderGroupsHandler(repo domain.RepositoryReaderGroup) ListReaderGroupsHandler {
	if repo == nil {
		panic("nil repo")
	}
	return ListReaderGroupsHandler{repo: repo}
}

func (h ListReaderGroupsHandler) Handle(ctx context.Context, q ListReaderGroups) ([]ReaderGroupDTO, error) {
	groups, err := h.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader groups: %w", err)
	}

	dtos := make([]ReaderGroupDTO, 0, len(groups))
	for _, group := range groups {
		dtos = append(dtos, ReaderGroupDTO{
			ID:             group.ID.String(),
			Name:           group.Name,
			StartOffset:    group.StartOffset,
			ReadersCount:   group.ReadersCount(),
			CalendarsCount: group.CalendarsCount(),
			CreatedAt:      group.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return dtos, nil
}
