package mqtthandler

import (
	"database/sql"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionGetCourts = debug.NewFunction(pkg, "GetCourts")
)

func GetCourts(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionGetCourts
	DebugVerbose(f, requestID, "")

	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	listOfCourts, err := model.ListCourtsTx(db)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	reply := struct {
		Status       int           `json:"status"`
		Message      string        `json:"message"`
		ListOfCourts []model.Court `json:"listOfCourts"`
	}{
		Status:       StatusOK,
		Message:      "ok",
		ListOfCourts: listOfCourts,
	}

	Reply(requestID, client, replyTopic, reply)
}
