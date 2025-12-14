package adapters

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/asdine/storm/v3"
	"github.com/gofrs/uuid/v5"
)

type CalendarRefDB struct {
	ID        string             `json:"id"`
	Year      int                `json:"year"`
	Calendar  domain.CalendarMap `json:"calendar"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

type ReaderGroupDB struct {
	ID          string            `storm:"id" json:"id"`
	Name        string            `storm:"index" json:"name"`
	Readers     []PsalmReaderTGDB `json:"readers"`
	StartOffset int               `json:"start_offset"`
	Calendars   []CalendarRefDB   `json:"calendars"`
	CreatedAt   time.Time         `storm:"index" json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ReaderGroupRepository struct {
	db *storm.DB
}

func NewReaderGroupRepository(db *storm.DB) *ReaderGroupRepository {
	if db == nil {
		slog.Error("missing db in NewReaderGroupRepository")
		os.Exit(1)
	}
	return &ReaderGroupRepository{db: db}
}

func (r *ReaderGroupRepository) Create(ctx context.Context, group *domain.ReaderGroup) error {
	dbGroup := r.marshalToDB(group)
	err := r.db.Save(&dbGroup)
	if err != nil {
		if errors.Is(err, storm.ErrAlreadyExists) {
			return fmt.Errorf("reader group already exists: %w", err)
		}
		return fmt.Errorf("error creating reader group: %w", err)
	}
	return nil
}

func (r *ReaderGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
	var dbGroup ReaderGroupDB
	err := r.db.One("ID", id.String(), &dbGroup)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, fmt.Errorf("reader group with ID %s not found", id)
		}
		return nil, fmt.Errorf("error getting reader group: %w", err)
	}

	return r.unmarshalFromDB(&dbGroup)
}

func (r *ReaderGroupRepository) GetAll(ctx context.Context) ([]domain.ReaderGroup, error) {
	var dbGroups []ReaderGroupDB
	err := r.db.All(&dbGroups)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return []domain.ReaderGroup{}, nil
		}
		return nil, fmt.Errorf("error getting all reader groups: %w", err)
	}

	groups := make([]domain.ReaderGroup, 0, len(dbGroups))
	for i := range dbGroups {
		group, errUnm := r.unmarshalFromDB(&dbGroups[i])
		if errUnm != nil {
			return nil, fmt.Errorf("error unmarshalling reader group: %w", errUnm)
		}
		groups = append(groups, *group)
	}

	return groups, nil
}

func (r *ReaderGroupRepository) Update(ctx context.Context, group *domain.ReaderGroup) error {
	dbGroup := r.marshalToDB(group)
	err := r.db.Update(&dbGroup)
	if err != nil {
		return fmt.Errorf("error updating reader group: %w", err)
	}
	return nil
}

func (r *ReaderGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var dbGroup ReaderGroupDB
	dbGroup.ID = id.String()
	err := r.db.DeleteStruct(&dbGroup)
	if err != nil {
		return fmt.Errorf("error deleting reader group: %w", err)
	}
	return nil
}

func (r *ReaderGroupRepository) marshalToDB(group *domain.ReaderGroup) ReaderGroupDB {
	readers := make([]PsalmReaderTGDB, 0, len(group.Readers))
	for _, reader := range group.Readers {
		readers = append(readers, PsalmReaderTGDB{
			ID:         reader.ID,
			Username:   reader.Username,
			TelegramID: reader.TelegramID,
			Phone:      reader.Phone,
			CreatedAt:  reader.CreatedAt,
			UpdatedAt:  reader.UpdatedAt,
		})
	}

	calendars := make([]CalendarRefDB, 0, len(group.Calendars))
	for _, calendar := range group.Calendars {
		calendars = append(calendars, CalendarRefDB{
			ID:        calendar.ID.String(),
			Year:      calendar.Year,
			Calendar:  calendar.Calendar,
			CreatedAt: calendar.CreatedAt.Format(time.RFC3339),
			UpdatedAt: calendar.UpdatedAt.Format(time.RFC3339),
		})
	}

	return ReaderGroupDB{
		ID:          group.ID.String(),
		Name:        group.Name,
		Readers:     readers,
		StartOffset: group.StartOffset,
		Calendars:   calendars,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

func (r *ReaderGroupRepository) unmarshalFromDB(dbGroup *ReaderGroupDB) (*domain.ReaderGroup, error) {
	id, err := uuid.FromString(dbGroup.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	readers := make([]domain.PsalmReader, 0, len(dbGroup.Readers))
	for _, dbReader := range dbGroup.Readers {

		readers = append(readers, *domain.UnmarshallPsalmReader(
			dbReader.ID,
			dbReader.ReaderNumber,
			dbReader.Username,
			dbReader.TelegramID,
			dbReader.Phone,
			dbReader.CreatedAt,
			dbReader.UpdatedAt,
		))
	}

	calendars := make([]domain.CalendarOfReader, 0, len(dbGroup.Calendars))
	for _, dbCalendar := range dbGroup.Calendars {
		calendarID, err := uuid.FromString(dbCalendar.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid calendar ID: %w", err)
		}

		createdAt, err := time.Parse(time.RFC3339, dbCalendar.CreatedAt)
		if err != nil {
			slog.Warn("failed to parse calendar created_at", "error", err)
			createdAt = time.Now()
		}

		updatedAt, err := time.Parse(time.RFC3339, dbCalendar.UpdatedAt)
		if err != nil {
			slog.Warn("failed to parse calendar updated_at", "error", err)
			updatedAt = time.Now()
		}

		calendars = append(calendars, *domain.UnmarshallCalendarOfReader(
			calendarID,
			dbCalendar.Year,
			dbCalendar.Calendar,
			createdAt,
			updatedAt,
		))
	}

	return domain.UnmarshallReaderGroup(
		id,
		dbGroup.Name,
		readers,
		dbGroup.StartOffset,
		calendars,
		dbGroup.CreatedAt,
		dbGroup.UpdatedAt,
	), nil
}
