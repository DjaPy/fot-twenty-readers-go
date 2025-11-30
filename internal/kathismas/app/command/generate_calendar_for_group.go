package command

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type GenerateCalendarForGroup struct {
	GroupID uuid.UUID
	Year    int
}

type CalendarGenerator interface {
	GenerateForGroup(group *domain.ReaderGroup, year int) (*bytes.Buffer, error)
}

type GenerateCalendarForGroupHandler struct {
	groupRepo domain.RepositoryReaderGroup
	generator CalendarGenerator
}

func NewGenerateCalendarForGroupHandler(
	groupRepo domain.RepositoryReaderGroup,
	generator CalendarGenerator,
) GenerateCalendarForGroupHandler {
	if groupRepo == nil {
		panic("nil groupRepo")
	}
	if generator == nil {
		panic("nil generator")
	}
	return GenerateCalendarForGroupHandler{
		groupRepo: groupRepo,
		generator: generator,
	}
}

func (h GenerateCalendarForGroupHandler) Handle(ctx context.Context, cmd GenerateCalendarForGroup) (*bytes.Buffer, error) {
	group, err := h.groupRepo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader group: %w", err)
	}

	year := cmd.Year
	if year == 0 {
		year = time.Now().Year()
	}

	buffer, err := h.generator.GenerateForGroup(group, year)
	if err != nil {
		return nil, fmt.Errorf("failed to generate calendar: %w", err)
	}

	calendar := domain.CalendarOfReader{
		ID:        uuid.Must(uuid.NewV7()),
		Calendar:  make(domain.CalendarMap),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := group.AddCalendar(calendar); err != nil {
		return nil, fmt.Errorf("failed to add calendar to group: %w", err)
	}

	if err := h.groupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update reader group: %w", err)
	}

	return buffer, nil
}
