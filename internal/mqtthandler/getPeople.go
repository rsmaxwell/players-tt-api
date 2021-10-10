package mqtthandler

import (
	"database/sql"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionGetPeople = debug.NewFunction(pkg, "GetPeople")

	filters = make(map[string]string)
)

func init() {
	filters[""] = ``
	filters["all"] = ``
	filters["players"] = `WHERE status = 'player'`
	filters["inactive"] = `WHERE status = 'inactive'`
	filters["suspended"] = `WHERE status = 'suspended'`
}

func GetPeople(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionGetPeople
	DebugVerbose(f, requestID, "")

	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	filter, err := GetStringFromRequest(f, requestID, "filter", data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	var whereClause string
	var ok bool
	if whereClause, ok = filters[filter]; !ok {
		ReplyBadRequest(requestID, client, replyTopic, fmt.Sprintf("unexpected filter name: '%s'", filter))
		return
	}

	listOfFullPeople, err := model.ListPeopleTx(db, whereClause)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	listOfPeople := []model.Person{}
	for _, person := range listOfFullPeople {
		listOfPeople = append(listOfPeople, *person.ToLimited())
	}

	reply := struct {
		Status       int            `json:"status"`
		Message      string         `json:"message"`
		ListOfPeople []model.Person `json:"listOfPeople"`
	}{
		Status:       StatusOK,
		Message:      "ok",
		ListOfPeople: listOfPeople,
	}

	Reply(requestID, client, replyTopic, reply)
}
