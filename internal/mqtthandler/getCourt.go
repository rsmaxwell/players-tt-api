package mqtthandler

import (
	"database/sql"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionGetCourt = debug.NewFunction(pkg, "GetCourt")
)

// GetCourt method
func GetCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionGetCourt
	DebugVerbose(f, requestID, "")

	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	id, err := GetIntegerFromRequest(f, requestID, "id", data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "ID: %d", id)

	c := model.Court{ID: id}
	err = c.LoadCourtTx(db)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	reply := struct {
		Status  int               `json:"status"`
		Message string            `json:"message"`
		Person  *model.PlainCourt `json:"court"`
	}{
		Status:  StatusOK,
		Message: "ok",
		Person:  c.ToPlainCourt(),
	}

	Reply(requestID, client, replyTopic, reply)
}
