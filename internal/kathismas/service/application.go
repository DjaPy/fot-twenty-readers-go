package service

import (
	"context"
	"log/slog"
	"os"

	"github.com/DjaPy/fot-twenty-readers-go/internal/common/metrics"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters/excel"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
)

func NewApplication(ctx context.Context, logger *slog.Logger) *app.Application {
	db, err := storm.Open("for-twenty-readers.db", storm.Codec(json.Codec))
	if err != nil {
		slog.Error("could not open database", "error", err)
		os.Exit(1)
	}

	cleanup := func() {
		if errClose := db.Close(); errClose != nil {
			slog.Error("could not close database", "error", errClose)
		}
	}

	metricsClient := metrics.NoOp{}

	psalmReaderTGRepository := adapters.NewPsalmReaderTGRepository(db)
	readerGroupRepository := adapters.NewReaderGroupRepository(db)
	calendarGenerator := excel.NewCalendarGenerator()

	return app.NewApplication(
		app.Commands{
			CreateCalendarOfReader:   command.NewCreatePsalmReaderTGHandler(psalmReaderTGRepository, logger, metricsClient),
			CreateReaderGroup:        command.NewCreateReaderGroupHandler(readerGroupRepository),
			AddReaderToGroup:         command.NewAddReaderToGroupHandler(readerGroupRepository),
			GenerateCalendarForGroup: command.NewGenerateCalendarForGroupHandler(readerGroupRepository, calendarGenerator),
		},
		app.Queries{
			ListReaderGroups:      query.NewListReaderGroupsHandler(readerGroupRepository),
			GetReaderGroup:        query.NewGetReaderGroupHandler(readerGroupRepository),
			GetCurrentKathisma:    query.NewGetCurrentKathismaHandler(readerGroupRepository),
			GetReaderByTelegramID: query.NewGetReaderByTelegramIDHandler(readerGroupRepository),
		},
		cleanup,
	)
}
