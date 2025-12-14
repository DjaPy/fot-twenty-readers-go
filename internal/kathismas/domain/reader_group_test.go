package domain

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReaderGroup(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		startOffset int
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid group",
			groupName:   "Храм Покрова",
			startOffset: 1,
			wantErr:     false,
		},
		{
			name:        "valid with max offset",
			groupName:   "Test Group",
			startOffset: 20,
			wantErr:     false,
		},
		{
			name:        "empty name",
			groupName:   "",
			startOffset: 1,
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "offset too small",
			groupName:   "Test Group",
			startOffset: 0,
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
		{
			name:        "offset too large",
			groupName:   "Test Group",
			startOffset: 21,
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
		{
			name:        "negative offset",
			groupName:   "Test Group",
			startOffset: -1,
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewReaderGroup(tt.groupName, tt.startOffset)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, group)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, group)
			assert.Equal(t, tt.groupName, group.Name)
			assert.Equal(t, tt.startOffset, group.StartOffset)
			assert.NotEqual(t, uuid.Nil, group.ID)
			assert.Empty(t, group.Readers)
			assert.Empty(t, group.Calendars)
			assert.False(t, group.CreatedAt.IsZero())
			assert.False(t, group.UpdatedAt.IsZero())
		})
	}
}

func TestReaderGroup_AddReader(t *testing.T) {
	tests := []struct {
		name          string
		setupGroup    func() *ReaderGroup
		readerNumber  int8
		username      string
		wantErr       bool
		errContains   string
		validateGroup func(t *testing.T, group *ReaderGroup)
	}{
		{
			name: "add valid reader",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				return group
			},
			readerNumber: 1,
			username:     "Иван",
			wantErr:      false,
			validateGroup: func(t *testing.T, group *ReaderGroup) {
				assert.Len(t, group.Readers, 1)
				assert.Equal(t, int8(1), group.Readers[0].ReaderNumber)
				assert.Equal(t, "Иван", group.Readers[0].Username)
			},
		},
		{
			name: "add multiple readers",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				reader1, _ := NewPsalmReader("Петр", 0, "", 1)
				_ = group.AddReader(reader1)
				return group
			},
			readerNumber: 2,
			username:     "Павел",
			wantErr:      false,
			validateGroup: func(t *testing.T, group *ReaderGroup) {
				assert.Len(t, group.Readers, 2)
			},
		},
		{
			name: "duplicate reader number",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				reader1, _ := NewPsalmReader("Иван", 0, "", 1)
				_ = group.AddReader(reader1)
				return group
			},
			readerNumber: 1,
			username:     "Петр",
			wantErr:      true,
			errContains:  "reader number 1 is already taken",
		},
		{
			name: "group at maximum capacity",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				for i := int8(1); i <= 20; i++ {
					reader, _ := NewPsalmReader("Reader", 0, "", i)
					_ = group.AddReader(reader)
				}
				return group
			},
			readerNumber: 21,
			username:     "Новый чтец",
			wantErr:      true,
			errContains:  "maximum number of readers",
		},
		{
			name: "duplicate telegram id",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				reader1, _ := NewPsalmReader("Иван", 123456, "", 1)
				_ = group.AddReader(reader1)
				return group
			},
			readerNumber: 2,
			username:     "Петр",
			wantErr:      true,
			errContains:  "telegram ID 123456 already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := tt.setupGroup()
			oldUpdatedAt := group.UpdatedAt

			var reader *PsalmReader
			var err error

			if tt.name == "duplicate telegram id" {
				reader, err = NewPsalmReader(tt.username, 123456, "", tt.readerNumber)
			} else {
				reader, err = NewPsalmReader(tt.username, 0, "", tt.readerNumber)
			}

			if err == nil {
				err = group.AddReader(reader)
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			assert.True(t, group.UpdatedAt.After(oldUpdatedAt), "UpdatedAt should be updated")

			if tt.validateGroup != nil {
				tt.validateGroup(t, group)
			}
		})
	}
}

func TestReaderGroup_RemoveReader(t *testing.T) {
	tests := []struct {
		name        string
		setupGroup  func() (*ReaderGroup, uuid.UUID)
		readerID    func(group *ReaderGroup, existingID uuid.UUID) uuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name: "remove existing reader",
			setupGroup: func() (*ReaderGroup, uuid.UUID) {
				group, _ := NewReaderGroup("Test", 1)
				reader, _ := NewPsalmReader("Иван", 0, "", 1)
				_ = group.AddReader(reader)
				return group, reader.ID
			},
			readerID: func(group *ReaderGroup, existingID uuid.UUID) uuid.UUID {
				return existingID
			},
			wantErr: false,
		},
		{
			name: "remove non-existent reader",
			setupGroup: func() (*ReaderGroup, uuid.UUID) {
				group, _ := NewReaderGroup("Test", 1)
				return group, uuid.Nil
			},
			readerID: func(group *ReaderGroup, existingID uuid.UUID) uuid.UUID {
				id, _ := uuid.NewV7()
				return id
			},
			wantErr:     true,
			errContains: "not found in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, existingID := tt.setupGroup()
			initialCount := len(group.Readers)
			readerID := tt.readerID(group, existingID)

			err := group.RemoveReader(readerID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Len(t, group.Readers, initialCount)
				return
			}

			require.NoError(t, err)
			assert.Len(t, group.Readers, initialCount-1)
		})
	}
}

func TestReaderGroup_UpdateName(t *testing.T) {
	tests := []struct {
		name        string
		newName     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid name update",
			newName: "Новое название",
			wantErr: false,
		},
		{
			name:        "empty name",
			newName:     "",
			wantErr:     true,
			errContains: "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, _ := NewReaderGroup("Old Name", 1)
			oldUpdatedAt := group.UpdatedAt

			time.Sleep(time.Millisecond)

			err := group.UpdateName(tt.newName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, "Old Name", group.Name)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.newName, group.Name)
			assert.True(t, group.UpdatedAt.After(oldUpdatedAt))
		})
	}
}

func TestReaderGroup_UpdateStartOffset(t *testing.T) {
	tests := []struct {
		name        string
		newOffset   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid offset 1",
			newOffset: 1,
			wantErr:   false,
		},
		{
			name:      "valid offset 20",
			newOffset: 20,
			wantErr:   false,
		},
		{
			name:        "offset too small",
			newOffset:   0,
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
		{
			name:        "offset too large",
			newOffset:   21,
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, _ := NewReaderGroup("Test", 10)
			oldUpdatedAt := group.UpdatedAt

			time.Sleep(time.Millisecond)

			err := group.UpdateStartOffset(tt.newOffset)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, 10, group.StartOffset)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.newOffset, group.StartOffset)
			assert.True(t, group.UpdatedAt.After(oldUpdatedAt))
		})
	}
}

func TestReaderGroup_GetAvailableReaderNumbers(t *testing.T) {
	tests := []struct {
		name           string
		setupGroup     func() *ReaderGroup
		expectedResult []int8
	}{
		{
			name: "empty group",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				return group
			},
			expectedResult: []int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			name: "group with some readers",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				reader1, _ := NewPsalmReader("Reader1", 0, "", 1)
				reader5, _ := NewPsalmReader("Reader5", 0, "", 5)
				reader10, _ := NewPsalmReader("Reader10", 0, "", 10)
				_ = group.AddReader(reader1)
				_ = group.AddReader(reader5)
				_ = group.AddReader(reader10)
				return group
			},
			expectedResult: []int8{2, 3, 4, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			name: "full group",
			setupGroup: func() *ReaderGroup {
				group, _ := NewReaderGroup("Test", 1)
				for i := int8(1); i <= 20; i++ {
					reader, _ := NewPsalmReader("Reader", 0, "", i)
					_ = group.AddReader(reader)
				}
				return group
			},
			expectedResult: []int8{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := tt.setupGroup()
			available := group.GetAvailableReaderNumbers()
			assert.Equal(t, tt.expectedResult, available)
		})
	}
}

func TestReaderGroup_IsReaderNumberAvailable(t *testing.T) {
	group, _ := NewReaderGroup("Test", 1)
	reader1, _ := NewPsalmReader("Reader1", 0, "", 1)
	_ = group.AddReader(reader1)

	tests := []struct {
		name     string
		number   int8
		expected bool
	}{
		{"taken number", 1, false},
		{"available number", 2, true},
		{"number below range", 0, false},
		{"number above range", 21, false},
		{"valid available number", 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := group.IsReaderNumberAvailable(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}
