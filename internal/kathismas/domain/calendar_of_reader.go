package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// CalendarMap stores calendar data for all readers in a group
// First key: reader number (1-20)
// Second key: year day (1-365/366)
// Value: kathisma number (1-20)
type CalendarMap map[int]map[int]int

type CalendarOfReader struct {
	ID          uuid.UUID `storm:"id"`
	Year        int
	StartOffset int
	Calendar    CalendarMap
	CreatedAt   time.Time `storm:"index"`
	UpdatedAt   time.Time
}

func UnmarshallCalendarOfReader(
	id uuid.UUID,
	year int,
	startOffset int,
	calendar CalendarMap,
	createdAt time.Time,
	updatedAt time.Time,
) *CalendarOfReader {
	return &CalendarOfReader{
		ID:          id,
		Year:        year,
		StartOffset: startOffset,
		Calendar:    calendar,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func NewCalendarOfReader(year, startOffset int, calendarData CalendarMap) *CalendarOfReader {
	now := time.Now()
	return &CalendarOfReader{
		ID:          uuid.Must(uuid.NewV7()),
		Year:        year,
		StartOffset: startOffset,
		Calendar:    calendarData,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CalculateNextStartOffset calculates the StartOffset for the next year
// based on the last kathisma of reader #1 in the current year
func (c *CalendarOfReader) CalculateNextStartOffset() int {
	readerOneSchedule, exists := c.Calendar[1]
	if !exists || len(readerOneSchedule) == 0 {
		return 1
	}

	maxDay := 0
	for day := range readerOneSchedule {
		if day > maxDay {
			maxDay = day
		}
	}

	lastKathisma := readerOneSchedule[maxDay]
	nextKathisma := lastKathisma + 1
	if nextKathisma > 20 {
		nextKathisma = 1
	}

	return nextKathisma
}
