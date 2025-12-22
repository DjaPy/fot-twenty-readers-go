package command

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/gofrs/uuid/v5"
)

type RegenerateCalendarForGroup struct {
	GroupID uuid.UUID
	Year    int
}

type RegenerateCalendarForGroupHandler struct {
	groupRepo domain.RepositoryReaderGroup
	generator CalendarGenerator
}

func NewRegenerateCalendarForGroupHandler(
	groupRepo domain.RepositoryReaderGroup,
	generator CalendarGenerator,
) RegenerateCalendarForGroupHandler {
	if groupRepo == nil || generator == nil {
		slog.Error("not found group repo or generator calendar in NewRegenerateCalendarForGroupHandler")
		os.Exit(1)
	}
	return RegenerateCalendarForGroupHandler{
		groupRepo: groupRepo,
		generator: generator,
	}
}

func (h RegenerateCalendarForGroupHandler) Handle(ctx context.Context, cmd RegenerateCalendarForGroup) (*bytes.Buffer, error) {
	group, err := h.groupRepo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader group: %w", err)
	}

	year := cmd.Year
	if year == 0 {
		year = time.Now().Year()
	}

	removed := group.RemoveCalendarsByYear(year)
	slog.Info("removed calendars for regeneration", "year", year, "count", removed)

	buffer, calendarData, err := h.generator.GenerateForGroup(group, year)
	if err != nil {
		return nil, fmt.Errorf("failed to generate calendar: %w", err)
	}

	calendar := domain.NewCalendarOfReader(year, calendarData)

	if err := group.AddCalendar(*calendar); err != nil {
		return nil, fmt.Errorf("failed to add calendar to group: %w", err)
	}

	if err := h.groupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update reader group: %w", err)
	}

	return buffer, nil
}
