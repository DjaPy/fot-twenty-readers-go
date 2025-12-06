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
	ID        uuid.UUID `storm:"id"`
	Year      int
	Calendar  CalendarMap
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time
}

func UnmarshallCalendarOfReader(
	id uuid.UUID,
	year int,
	calendar CalendarMap,
	createdAt time.Time,
	updatedAt time.Time,
) *CalendarOfReader {
	return &CalendarOfReader{
		ID:        id,
		Year:      year,
		Calendar:  calendar,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func NewCalendarOfReader(year int, calendarData CalendarMap) *CalendarOfReader {
	now := time.Now()
	return &CalendarOfReader{
		ID:        uuid.Must(uuid.NewV7()),
		Year:      year,
		Calendar:  calendarData,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
