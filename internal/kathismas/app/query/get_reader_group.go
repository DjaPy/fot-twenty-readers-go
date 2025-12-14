package query

import (
	"context"
	"fmt"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type GetReaderGroup struct {
	ID uuid.UUID
}

type PsalmReaderDTO struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	ReaderNumber int8   `json:"reader_number"`
	TelegramID   int64  `json:"telegram_id"`
	Phone        string `json:"phone"`
}

type ReaderGroupDetailDTO struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	StartOffset int              `json:"start_offset"`
	Readers     []PsalmReaderDTO `json:"readers"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

func (dto *ReaderGroupDetailDTO) GetAvailableReaderNumbers() []int8 {
	usedNumbers := make(map[int8]bool)
	for _, r := range dto.Readers {
		usedNumbers[r.ReaderNumber] = true
	}

	available := make([]int8, 0, 20-len(dto.Readers))
	for i := int8(1); i <= 20; i++ {
		if !usedNumbers[i] {
			available = append(available, i)
		}
	}
	return available
}

type GetReaderGroupHandler struct {
	repo domain.RepositoryReaderGroup
}

func NewGetReaderGroupHandler(repo domain.RepositoryReaderGroup) GetReaderGroupHandler {
	if repo == nil {
		panic("nil repo")
	}
	return GetReaderGroupHandler{repo: repo}
}

func (h GetReaderGroupHandler) Handle(ctx context.Context, q GetReaderGroup) (*ReaderGroupDetailDTO, error) {
	group, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader group: %w", err)
	}

	readers := make([]PsalmReaderDTO, 0, len(group.Readers))
	for _, reader := range group.Readers {
		readers = append(readers, PsalmReaderDTO{
			ID:           reader.ID.String(),
			Username:     reader.Username,
			ReaderNumber: reader.ReaderNumber,
			TelegramID:   reader.TelegramID,
			Phone:        reader.Phone,
		})
	}

	return &ReaderGroupDetailDTO{
		ID:          group.ID.String(),
		Name:        group.Name,
		StartOffset: group.StartOffset,
		Readers:     readers,
		CreatedAt:   group.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
