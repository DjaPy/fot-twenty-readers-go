package service

import (
	"context"
	"github.com/DjaPy/fot-twenty-readers-go/internal/common/metrics"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
	"github.com/sirupsen/logrus"
	"log"
)

func NewApplication(ctx context.Context) *app.Application {
	db, err := storm.Open("for-twenty-readers.db", storm.Codec(json.Codec))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			log.Printf("Could not close database: %v", err)
		}
	}

	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.NoOp{}

	psalmReaderTGRepository := adapters.NewPsalmReaderTGRepository(db)
	return app.NewApplication(
		app.Commands{
			CreateCalendarOfReader: command.NewCreatePsalmReaderTGHandler(psalmReaderTGRepository, logger, metricsClient),
		},
		app.Queries{},
		cleanup,
	)
}
