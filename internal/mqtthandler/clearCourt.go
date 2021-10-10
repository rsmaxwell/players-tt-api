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
	functionClearCourt = debug.NewFunction(pkg, "ClearCourt")
)

// ClearCourt method
func ClearCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionClearCourt
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

	err = model.ClearCourtTx(db, courtID)
	if err != nil {
		message := "problem clearing court"
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

	ReplyOK(requestID, client, replyTopic)
}
