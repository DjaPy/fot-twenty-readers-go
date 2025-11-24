package adapters

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/asdine/storm/v3"
	"github.com/gofrs/uuid/v5"
)

type CalendarOfReaderDB struct {
	ID        uuid.UUID `storm:"id"`
	Calendar  domain.CalendarMap
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time
}

type CalendarOfReaderRepository struct {
	db *storm.DB
}

func NewCalendarOfReaderRepository(db *storm.DB) *CalendarOfReaderRepository {
	if db == nil {
		log.Fatal("missing db")
	}
	return &CalendarOfReaderRepository{db: db}
}

func (cr CalendarOfReaderRepository) GetCalendar(id uuid.UUID) (*domain.CalendarOfReader, error) {
	var CalendarOfReaderFromDB CalendarOfReaderDB
	err := cr.db.One("ID", id, &CalendarOfReaderFromDB)
	if err != nil {
		return nil, fmt.Errorf("getting calendar by id %v", err)
	}
	CalendarOfReader := domain.UnmarshallCalendarOfReader(
		CalendarOfReaderFromDB.ID,
		CalendarOfReaderFromDB.Calendar,
		CalendarOfReaderFromDB.CreatedAt,
		CalendarOfReaderFromDB.UpdatedAt,
	)
	return CalendarOfReader, nil
}

func (cr CalendarOfReaderRepository) CreateCalendarOfReader(
	calendarOfReader *domain.CalendarOfReader,
) error {
	err := cr.db.Save(&calendarOfReader)
	if err != nil {
		if errors.Is(err, storm.ErrAlreadyExists) {
			return fmt.Errorf("failed created calendar of reader")
		}
	}
	return nil
}
