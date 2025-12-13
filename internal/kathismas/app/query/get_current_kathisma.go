package query

import (
	"context"
	"fmt"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type GetCurrentKathisma struct {
	GroupID      uuid.UUID
	ReaderNumber int
}

type CurrentKathismaDTO struct {
	GroupID      uuid.UUID `json:"group_id"`
	GroupName    string    `json:"group_name"`
	ReaderNumber int       `json:"reader_number"`
	Date         string    `json:"date"`
	YearDay      int       `json:"year_day"`
	Kathisma     int       `json:"kathisma"`
	Year         int       `json:"year"`
}

type GetCurrentKathismaHandler struct {
	groupRepo domain.RepositoryReaderGroup
}

func NewGetCurrentKathismaHandler(groupRepo domain.RepositoryReaderGroup) GetCurrentKathismaHandler {
	return GetCurrentKathismaHandler{groupRepo: groupRepo}
}

func (h GetCurrentKathismaHandler) Handle(ctx context.Context, query GetCurrentKathisma) (*CurrentKathismaDTO, error) {
	if query.ReaderNumber < 1 || query.ReaderNumber > 20 {
		return nil, fmt.Errorf("reader number must be between 1 and 20")
	}

	group, err := h.groupRepo.GetByID(ctx, query.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	now := time.Now()
	currentYear := now.Year()
	yearDay := now.YearDay()

	var currentCalendar *domain.CalendarOfReader
	for _, cal := range group.Calendars {
		if cal.Year == currentYear {
			currentCalendar = &cal
			break
		}
	}

	if currentCalendar == nil {
		return nil, fmt.Errorf("no calendar found for year %d. Please generate calendar for this year first", currentYear)
	}

	readerCalendar, ok := currentCalendar.Calendar[query.ReaderNumber]
	if !ok {
		return nil, fmt.Errorf("reader number %d not found in calendar", query.ReaderNumber)
	}

	kathisma, ok := readerCalendar[yearDay]
	if !ok {
		return &CurrentKathismaDTO{
			GroupID:      group.ID,
			GroupName:    group.Name,
			ReaderNumber: query.ReaderNumber,
			Date:         now.Format("2006-01-02"),
			YearDay:      yearDay,
			Kathisma:     0, // 0 means no reading today
			Year:         currentYear,
		}, nil
	}

	return &CurrentKathismaDTO{
		GroupID:      group.ID,
		GroupName:    group.Name,
		ReaderNumber: query.ReaderNumber,
		Date:         now.Format("2006-01-02"),
		YearDay:      yearDay,
		Kathisma:     kathisma,
		Year:         currentYear,
	}, nil
}
