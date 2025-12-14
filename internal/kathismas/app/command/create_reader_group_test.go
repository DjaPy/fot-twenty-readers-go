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

func TestCreateReaderGroupHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CreateReaderGroup
		setupMock   func(repo *mocks.RepositoryReaderGroupMock)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock)
	}{
		{
			name: "successful creation",
			cmd: CreateReaderGroup{
				Name:        "Храм Покрова",
				StartOffset: 1,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.CreateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					require.NotNil(t, group)
					assert.Equal(t, "Храм Покрова", group.Name)
					assert.Equal(t, 1, group.StartOffset)
					return nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock) {
				assert.NotEqual(t, uuid.Nil, groupID)
				assert.Len(t, repo.CreateCalls(), 1)
			},
		},
		{
			name: "creation with max offset",
			cmd: CreateReaderGroup{
				Name:        "Test Group",
				StartOffset: 20,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.CreateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					return nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.CreateCalls(), 1)
				call := repo.CreateCalls()[0]
				assert.Equal(t, 20, call.Group.StartOffset)
			},
		},
		{
			name: "empty name",
			cmd: CreateReaderGroup{
				Name:        "",
				StartOffset: 1,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				// Should not be called
			},
			wantErr:     true,
			errContains: "name cannot be empty",
			validate: func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock) {
				assert.Equal(t, uuid.Nil, groupID)
				assert.Len(t, repo.CreateCalls(), 0)
			},
		},
		{
			name: "invalid start offset - too small",
			cmd: CreateReaderGroup{
				Name:        "Test",
				StartOffset: 0,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				// Should not be called
			},
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
			validate: func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.CreateCalls(), 0)
			},
		},
		{
			name: "invalid start offset - too large",
			cmd: CreateReaderGroup{
				Name:        "Test",
				StartOffset: 21,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				// Should not be called
			},
			wantErr:     true,
			errContains: "start offset must be between 1 and 20",
		},
		{
			name: "repository error",
			cmd: CreateReaderGroup{
				Name:        "Test Group",
				StartOffset: 1,
			},
			setupMock: func(repo *mocks.RepositoryReaderGroupMock) {
				repo.CreateFunc = func(ctx context.Context, group *domain.ReaderGroup) error {
					return errors.New("database connection failed")
				}
			},
			wantErr:     true,
			errContains: "database connection failed",
			validate: func(t *testing.T, groupID uuid.UUID, repo *mocks.RepositoryReaderGroupMock) {
				assert.Len(t, repo.CreateCalls(), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repoMock := &mocks.RepositoryReaderGroupMock{}
			tt.setupMock(repoMock)
			handler := NewCreateReaderGroupHandler(repoMock)

			// Act
			groupID, err := handler.Handle(context.Background(), tt.cmd)

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
				tt.validate(t, groupID, repoMock)
			}
		})
	}
}
