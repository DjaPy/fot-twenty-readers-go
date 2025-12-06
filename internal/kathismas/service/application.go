package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/DjaPy/fot-twenty-readers-go/internal/common/metrics"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters/excel"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) *app.Application {
	db, err := storm.Open("for-twenty-readers.db", storm.Codec(json.Codec))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}

	cleanup := func() {
		if errClose := db.Close(); errClose != nil {
			slog.Error(fmt.Sprintf("Could not close database: %v", errClose))
		}
	}

	logger := logrus.NewEntry(logrus.StandardLogger())
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
			ListReaderGroups: query.NewListReaderGroupsHandler(readerGroupRepository),
			GetReaderGroup:   query.NewGetReaderGroupHandler(readerGroupRepository),
		},
		cleanup,
	)
}
