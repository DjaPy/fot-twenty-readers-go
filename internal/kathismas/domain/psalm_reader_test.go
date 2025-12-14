package domain

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPsalmReader(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		telegramID   int64
		phone        string
		readerNumber int8
	}{
		{
			name:         "valid reader with all fields",
			username:     "Иван Петров",
			telegramID:   123456789,
			phone:        "+79001234567",
			readerNumber: 1,
		},
		{
			name:         "valid reader without telegram and phone",
			username:     "Петр Сидоров",
			telegramID:   0,
			phone:        "",
			readerNumber: 5,
		},
		{
			name:         "valid reader with max number",
			username:     "Test Reader",
			telegramID:   0,
			phone:        "",
			readerNumber: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewPsalmReader(tt.username, tt.telegramID, tt.phone, tt.readerNumber)

			require.NoError(t, err)
			require.NotNil(t, reader)
			assert.Equal(t, tt.username, reader.Username)
			assert.Equal(t, tt.telegramID, reader.TelegramID)
			assert.Equal(t, tt.phone, reader.Phone)
			assert.Equal(t, tt.readerNumber, reader.ReaderNumber)
			assert.NotEqual(t, uuid.Nil, reader.ID)
			assert.False(t, reader.CreatedAt.IsZero())
			assert.False(t, reader.UpdatedAt.IsZero())
		})
	}
}
