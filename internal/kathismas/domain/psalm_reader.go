package domain

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
)

type PsalmReader struct {
	ID           uuid.UUID
	ReaderNumber int8
	Username     string
	TelegramID   int64
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewPsalmReader(username string, telegramID int64, phone string, readerNumber int8) (*PsalmReader, error) {
	ID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed generate uuid7 %v", err)
	}
	createdAt := time.Now()
	updatedAt := time.Now()
	return &PsalmReader{
		ID:           ID,
		ReaderNumber: readerNumber,
		Username:     username,
		TelegramID:   telegramID,
		Phone:        phone,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func UnmarshallPsalmReader(
	id uuid.UUID,
	readerNumber int8,
	username string,
	telegramID int64,
	phone string,
	createdAt time.Time,
	updatedAt time.Time,
) *PsalmReader {
	return &PsalmReader{
		ID:           id,
		ReaderNumber: readerNumber,
		Username:     username,
		TelegramID:   telegramID,
		Phone:        phone,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
