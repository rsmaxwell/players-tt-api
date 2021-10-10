package mqtthandler

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
func GetWaiters(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionGetWaiters
	DebugVerbose(f, requestID, "")

	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	waiters, err := model.ListWaiters(context.Background(), db)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	var list []model.DisplayWaiter
	for _, waiter := range waiters {

		p := model.FullPerson{ID: waiter.Person}
		err := p.LoadPerson(context.Background(), db)
		if err != nil {
			ReplyInternalServerError(requestID, client, replyTopic, err.Error())
			return
		}

		w := model.DisplayWaiter{}
		w.PersonID = waiter.Person
		w.Knownas = p.Knownas
		w.Start = waiter.Start.Unix()

		list = append(list, w)
	}

	reply := struct {
		Status        int                   `json:"status"`
		Message       string                `json:"message"`
		ListOfWaiters []model.DisplayWaiter `json:"listOfWaiters"`
	}{
		Status:        StatusOK,
		Message:       "ok",
		ListOfWaiters: list,
	}

	Reply(requestID, client, replyTopic, reply)
}
