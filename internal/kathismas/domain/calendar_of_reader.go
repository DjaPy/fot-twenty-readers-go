package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CalendarMap map[string]map[string]string

type CalendarOfReader struct {
	ID        uuid.UUID `storm:"id"`
	Calendar  CalendarMap
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time
}

func UnmarshallCalendarOfReader(
	id uuid.UUID,
	calendar CalendarMap,
	createdAt time.Time,
	updatedAt time.Time,
) *CalendarOfReader {
	return &CalendarOfReader{
		ID:        id,
		Calendar:  calendar,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
