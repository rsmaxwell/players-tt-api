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
	functionCreateCourt = debug.NewFunction(pkg, "CreateCourt")
)

// CreateCourt method
func CreateCourt(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionCreateCourt
	DebugVerbose(f, requestID, "")

	userID, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

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

	c, err := model.NewCourtFromMap(data)
	if err != nil {
		message := err.Error()
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
		return
	}

	err = c.SaveCourtTx(db)
	if err != nil {
		message := err.Error()
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
		return
	}

	ReplyOK(requestID, client, replyTopic)
}
