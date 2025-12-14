package command

import (
	"context"
	"errors"
	"testing"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain/mocks"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddReaderToGroupHandler_Handle(t *testing.T) {
	groupID, _ := uuid.NewV7()

	tests := []struct {
		name        string
		cmd         AddReaderToGroup
		setupMock   func(repo *mocks.RepositoryReaderGroupMock)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, repo *mocks.RepositoryReaderGroupMock)
	}{
		{
			name: "successful add reader",
			cmd: AddReaderToGroup{
				GroupID:      groupID,
				ReaderNumber: 1,
				Username:     "Иван Петров",
				TelegramID:   123456,
				Phone:        "+79001234567",
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
					group, _ := domain.NewReaderGroup("Test Group", 1)
					group.ID = groupID
					return group, nil
				}
				repo.UpdateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					assert.Len(t, group.Readers, 1)
					assert.Equal(t, "Иван Петров", group.Readers[0].Username)
					assert.Equal(t, int8(1), group.Readers[0].ReaderNumber)
					return nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.GetByIDCalls(), 1)
				assert.Len(t, repo.UpdateCalls(), 1)
			},
		},
		{
			name: "add second reader",
			cmd: AddReaderToGroup{
				GroupID:      groupID,
				ReaderNumber: 2,
				Username:     "Петр Сидоров",
				TelegramID:   0,
				Phone:        "",
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
					group, _ := domain.NewReaderGroup("Test Group", 1)
					group.ID = groupID
					// Add first reader
					reader1, _ := domain.NewPsalmReader("Иван", 0, "", 1)
					_ = group.AddReader(reader1)
					return group, nil
				}
				repo.UpdateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					assert.Len(t, group.Readers, 2)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "group not found",
			cmd: AddReaderToGroup{
				GroupID:      groupID,
				ReaderNumber: 1,
				Username:     "Test User",
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
					return nil, errors.New("group not found")
				}
			},
			wantErr:     true,
			errContains: "failed to get reader group",
			validate: func(t *testing.T, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.GetByIDCalls(), 1)
				assert.Len(t, repo.UpdateCalls(), 0)
			},
		},
		{
			name: "duplicate reader number",
			cmd: AddReaderToGroup{
				GroupID:      groupID,
				ReaderNumber: 1,
				Username:     "Duplicate User",
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
					group, _ := domain.NewReaderGroup("Test Group", 1)
					// Add reader with number 1
					reader1, _ := domain.NewPsalmReader("Existing User", 0, "", 1)
					_ = group.AddReader(reader1)
					return group, nil
				}
			},
			wantErr:     true,
			errContains: "reader number 1 is already taken",
			validate: func(t *testing.T, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.UpdateCalls(), 0)
			},
		},
		{
			name: "repository update error",
			cmd: AddReaderToGroup{
				GroupID:      groupID,
				ReaderNumber: 1,
				Username:     "Test User",
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.ReaderGroup, error) {
					group, _ := domain.NewReaderGroup("Test Group", 1)
					return group, nil
				}
				repo.UpdateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					return errors.New("database write failed")
				}
			},
			wantErr:     true,
			errContains: "database write failed",
			validate: func(t *testing.T, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.UpdateCalls(), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repoMock := &mocks.RepositoryReaderGroupMock{}
			tt.setupMock(repoMock)
			handler := NewAddReaderToGroupHandler(repoMock)

			// Act
			err := handler.Handle(context.Background(), tt.cmd)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, repoMock)
			}
		})
	}
}
