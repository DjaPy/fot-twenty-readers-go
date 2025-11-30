package app

import (
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
	cleanup  func()
}

func NewApplication(commands Commands, queries Queries, cleanup func()) *Application {
	return &Application{
		Commands: commands,
		Queries:  queries,
		cleanup:  cleanup,
	}
}

func (a *Application) Close() {
	if a.cleanup != nil {
		a.cleanup()
	}
}

type Commands struct {
	CreateCalendarOfReader   command.CreateCalendarOfReaderHandler
	CreateReaderGroup        command.CreateReaderGroupHandler
	AddReaderToGroup         command.AddReaderToGroupHandler
	GenerateCalendarForGroup command.GenerateCalendarForGroupHandler
}

type Queries struct {
	ListReaderGroups query.ListReaderGroupsHandler
	GetReaderGroup   query.GetReaderGroupHandler
}
