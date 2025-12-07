package adapters

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/asdine/storm/v3"
	"github.com/gofrs/uuid/v5"
)

type PsalmReaderTGDB struct {
	ID         uuid.UUID `storm:"id"`
	Username   string    `storm:"index"`
	TelegramID int64     `storm:"index, unique"`
	Phone      string
	CreatedAt  time.Time `storm:"index"`
	UpdatedAt  time.Time
}

type PsalmReaderTGRepository struct {
	db *storm.DB
}

func NewPsalmReaderTGRepository(db *storm.DB) *PsalmReaderTGRepository {
	if db == nil {
		log.Fatal("missing db")
	}
	return &PsalmReaderTGRepository{db: db}
}

func (pr PsalmReaderTGRepository) GetPsalmReaderTG(ctx context.Context, id uuid.UUID) (*domain.PsalmReader, error) {
	var dbPsalmReaderTG PsalmReaderTGDB
	err := pr.db.One("ID", id, &dbPsalmReaderTG)
	if err != nil {
		return nil, fmt.Errorf("error getting psalm reader: %v", err)
	}

	psalmReaderTG := domain.UnmarshallPsalmReader(
		dbPsalmReaderTG.ID,
		dbPsalmReaderTG.Username,
		dbPsalmReaderTG.TelegramID,
		dbPsalmReaderTG.Phone,
		dbPsalmReaderTG.CreatedAt,
		dbPsalmReaderTG.UpdatedAt,
	)
	return psalmReaderTG, nil
}

func (pr PsalmReaderTGRepository) CreatePsalmReaderTG(ctx context.Context, psalmReader *domain.PsalmReader) error {
	err := pr.db.Save(&psalmReader)
	if err != nil {
		if errors.Is(err, storm.ErrAlreadyExists) {
			return fmt.Errorf("psalm reader already exists: %v", err)
		}
	}
	return nil
}
