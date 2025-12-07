package domain

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

type NotFoundError struct {
	PsalmReaderUUID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("PsalmReader '%s' not found", e.PsalmReaderUUID)
}

type RepositoryPsalmReader interface {
	GetPsalmReaderTG(ctx context.Context, id uuid.UUID) (*PsalmReader, error)
	CreatePsalmReaderTG(ctx context.Context, psalmReader *PsalmReader) error
}

type RepositoryCalendarOfReaders interface {
	GetCalendar(id uuid.UUID) (*CalendarOfReader, error)
	CreateCalendarOfReader(calendarOfReader *CalendarOfReader) error
}

type RepositoryReaderGroup interface {
	Create(ctx context.Context, group *ReaderGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReaderGroup, error)
	GetAll(ctx context.Context) ([]ReaderGroup, error)
	Update(ctx context.Context, group *ReaderGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
}
