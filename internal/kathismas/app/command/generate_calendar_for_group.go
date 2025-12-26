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

type GenerateCalendarForGroup struct {
	GroupID     uuid.UUID
	Year        int
	StartOffset int
}

type CalendarGenerator interface {
	GenerateForGroup(year, startOffset int) (*bytes.Buffer, domain.CalendarMap, error)
}

type GenerateCalendarForGroupHandler struct {
	groupRepo domain.RepositoryReaderGroup
	generator CalendarGenerator
}

func NewGenerateCalendarForGroupHandler(
	groupRepo domain.RepositoryReaderGroup,
	generator CalendarGenerator,
) GenerateCalendarForGroupHandler {
	if groupRepo == nil || generator == nil {
		slog.Error("not found group repo or generator calendar in NewGenerateCalendarForGroupHandler")
		os.Exit(1)
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
	startOffset := h.calculateStartOffset(group, year, cmd.StartOffset)

	buffer, calendarData, err := h.generator.GenerateForGroup(year, startOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to generate calendar: %w", err)
	}

	calendar := domain.NewCalendarOfReader(year, startOffset, calendarData)

	if err := group.AddCalendar(*calendar); err != nil {
		return nil, fmt.Errorf("failed to add calendar to group: %w", err)
	}

	if err := h.groupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update reader group: %w", err)
	}

	return buffer, nil
}

func (h GenerateCalendarForGroupHandler) calculateStartOffset(group *domain.ReaderGroup, year, cmdStartOffset int) int {
	if cmdStartOffset != 0 {
		return cmdStartOffset
	}
	for _, cal := range group.Calendars {
		if cal.Year == year-1 {
			return cal.CalculateNextStartOffset()
		}
	}
	return group.StartOffset
}
