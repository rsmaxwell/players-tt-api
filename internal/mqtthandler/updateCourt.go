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
	functionUpdateCourt = debug.NewFunction(pkg, "UpdateCourt")
)

// UpdateCourt method
func UpdateCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionUpdateCourt
	DebugVerbose(f, requestID, "")

	userID, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	courtID, err := GetIntegerFromRequest(f, requestID, "id", data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "courtID: %d", courtID)

	user := model.FullPerson{ID: userID}
	err = user.LoadPersonTx(db)
	if err != nil {
		message := fmt.Sprintf("Could not load person [%d]", userID)
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
		return
	}

	err = user.CanEditCourt()
	if err != nil {
		message := fmt.Sprintf("Person [%d] is not allowed to edit court", userID)
		DebugVerbose(f, requestID, message)
		ReplyForbidden(requestID, client, replyTopic, message)
		return
	}

	err = model.UpdateCourtFieldsTx(db, courtID, *data)
	if err != nil {
		message := fmt.Sprintf("problem updating court fields: courtID: %d", courtID)
		DebugVerbose(f, requestID, message)
		ReplyForbidden(requestID, client, replyTopic, message)
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
