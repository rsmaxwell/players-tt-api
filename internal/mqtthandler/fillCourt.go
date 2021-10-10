package mqtthandler

import (
	"database/sql"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/publisher"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionFillCourt = debug.NewFunction(pkg, "FillCourt")
)

// FillCourt method
func FillCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionFillCourt
	DebugVerbose(f, requestID, "")

	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	courtID, err := GetIntegerFromRequest(f, requestID, "courtID", data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "courtID: %d", courtID)

	positions, err := model.FillCourtTx(db, courtID)
	if err != nil {
		message := "problem filling court"
		d := Dump(f, requestID, message)
		d.AddString("courtID", fmt.Sprintf("%d", courtID))
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	err = publisher.UpdatePublications(db, client, cfg)
	if err != nil {
		message := err.Error()
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
	}

	reply := struct {
		Status    int              `json:"status"`
		Message   string           `json:"message"`
		Positions []model.Position `json:"positions"`
	}{
		Status:    StatusOK,
		Message:   "ok",
		Positions: positions,
	}

	Reply(requestID, client, replyTopic, reply)
}
