package domain

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
)

type ReaderGroup struct {
	ID          uuid.UUID
	Name        string
	Readers     []PsalmReader
	StartOffset int
	Calendars   []CalendarOfReader
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewReaderGroup(name string, startOffset int) (*ReaderGroup, error) {
	if err := validateReaderGroupParams(name, startOffset); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid7: %w", err)
	}

	now := time.Now()

	return &ReaderGroup{
		ID:          id,
		Name:        name,
		Readers:     make([]PsalmReader, 0, 20),
		StartOffset: startOffset,
		Calendars:   make([]CalendarOfReader, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func UnmarshallReaderGroup(
	id uuid.UUID,
	name string,
	readers []PsalmReader,
	startOffset int,
	calendars []CalendarOfReader,
	createdAt time.Time,
	updatedAt time.Time,
) *ReaderGroup {
	return &ReaderGroup{
		ID:          id,
		Name:        name,
		Readers:     readers,
		StartOffset: startOffset,
		Calendars:   calendars,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (rg *ReaderGroup) AddReader(reader *PsalmReader) error {
	if len(rg.Readers) >= 20 {
		return fmt.Errorf("group already has maximum number of readers (20)")
	}

	for _, r := range rg.Readers {
		if r.ID == reader.ID {
			return fmt.Errorf("reader with ID %s already exists in group", reader.ID)
		}
		if r.TelegramID == reader.TelegramID && reader.TelegramID != 0 {
			return fmt.Errorf("reader with telegram ID %d already exists in group", reader.TelegramID)
		}
		if r.ReaderNumber == reader.ReaderNumber {
			return fmt.Errorf("reader number %d is already taken in this group", reader.ReaderNumber)
		}
	}

	rg.Readers = append(rg.Readers, *reader)
	rg.UpdatedAt = time.Now()
	return nil
}

func (rg *ReaderGroup) RemoveReader(readerID uuid.UUID) error {
	for i, reader := range rg.Readers {
		if reader.ID == readerID {
			rg.Readers = append(rg.Readers[:i], rg.Readers[i+1:]...)
			rg.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("reader with ID %s not found in group", readerID)
}

func (rg *ReaderGroup) UpdateReader(updatedReader PsalmReader) error {
	for i, reader := range rg.Readers {
		if reader.ID == updatedReader.ID {
			rg.Readers[i] = updatedReader
			rg.Readers[i].UpdatedAt = time.Now()
			rg.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("reader with ID %s not found in group", updatedReader.ID)
}

func (rg *ReaderGroup) GetReader(readerID uuid.UUID) (*PsalmReader, error) {
	for _, reader := range rg.Readers {
		if reader.ID == readerID {
			return &reader, nil
		}
	}
	return nil, fmt.Errorf("reader with ID %s not found in group", readerID)
}

func (rg *ReaderGroup) AddCalendar(calendar CalendarOfReader) error {
	for _, c := range rg.Calendars {
		if c.ID == calendar.ID {
			return fmt.Errorf("calendar with ID %s already exists in group", calendar.ID)
		}
	}

	rg.Calendars = append(rg.Calendars, calendar)
	rg.UpdatedAt = time.Now()
	return nil
}

func (rg *ReaderGroup) GetLatestCalendar() (*CalendarOfReader, error) {
	if len(rg.Calendars) == 0 {
		return nil, fmt.Errorf("no calendars found in group")
	}

	latest := &rg.Calendars[0]
	for i := range rg.Calendars {
		if rg.Calendars[i].CreatedAt.After(latest.CreatedAt) {
			latest = &rg.Calendars[i]
		}
	}
	return latest, nil
}

func (rg *ReaderGroup) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	rg.Name = name
	rg.UpdatedAt = time.Now()
	return nil
}

func (rg *ReaderGroup) UpdateStartOffset(startOffset int) error {
	if startOffset < 1 || startOffset > 20 {
		return fmt.Errorf("start offset must be between 1 and 20")
	}
	rg.StartOffset = startOffset
	rg.UpdatedAt = time.Now()
	return nil
}

func (rg *ReaderGroup) ReadersCount() int {
	return len(rg.Readers)
}

func (rg *ReaderGroup) CalendarsCount() int {
	return len(rg.Calendars)
}

func (rg *ReaderGroup) GetAvailableReaderNumbers() []int8 {
	usedNumbers := make(map[int8]bool)
	for _, r := range rg.Readers {
		usedNumbers[r.ReaderNumber] = true
	}

	available := make([]int8, 0, 20-len(rg.Readers))
	for i := int8(1); i <= 20; i++ {
		if !usedNumbers[i] {
			available = append(available, i)
		}
	}
	return available
}

func (rg *ReaderGroup) IsReaderNumberAvailable(number int8) bool {
	if number < 1 || number > 20 {
		return false
	}
	for _, r := range rg.Readers {
		if r.ReaderNumber == number {
			return false
		}
	}
	return true
}

func validateReaderGroupParams(name string, startOffset int) error {
	if name == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	if startOffset < 1 || startOffset > 20 {
		return fmt.Errorf("start offset must be between 1 and 20, got %d", startOffset)
	}
	return nil
}
