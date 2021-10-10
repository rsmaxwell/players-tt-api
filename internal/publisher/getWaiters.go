package publisher

import (
	"context"
	"database/sql"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionGetWaiters = debug.NewFunction(pkg, "GetWaiters")
)

// GetWaiters method
func GetWaiters(db *sql.DB, client mqtt.Client, cfg *config.Config) ([]Entry, error) {
	f := functionGetWaiters
	f.DebugVerbose("")

	waiters, err := model.ListWaiters(context.Background(), db)
	if err != nil {
		f.DebugVerbose(err.Error())
		return nil, err
	}

	var listOfWaiters []model.DisplayWaiter
	for _, waiter := range waiters {

		p := model.FullPerson{ID: waiter.Person}
		err := p.LoadPerson(context.Background(), db)
		if err != nil {
			f.DebugVerbose(err.Error())
			return nil, err
		}

		w := model.DisplayWaiter{}
		w.PersonID = p.ID
		w.Knownas = p.Knownas
		w.Start = waiter.Start.Unix()

		listOfWaiters = append(listOfWaiters, w)
	}

	entry := Entry{topic: "getWaiters", object: listOfWaiters}
	array := []Entry{entry}
	return array, nil
}
