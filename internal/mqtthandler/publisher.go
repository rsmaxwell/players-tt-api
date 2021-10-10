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
	functionUpdatePublications = debug.NewFunction(pkg, "UpdatePublications")
)

// UpdatePublications method
func UpdatePublications(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionUpdatePublications
	DebugVerbose(f, requestID, "")

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

	ReplyOK(requestID, client, replyTopic)
}
