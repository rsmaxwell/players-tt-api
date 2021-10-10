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
	functionDeleteCourt = debug.NewFunction(pkg, "DeleteCourt")
)

// DeleteCourt method
func DeleteCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionDeleteCourt
	DebugVerbose(f, requestID, "")

	userID, err := checkAuthenticated(requestID, data)
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

	c := model.Court{ID: id}
	err = c.DeleteCourtTx(db)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	ReplyOK(requestID, client, replyTopic)
}
