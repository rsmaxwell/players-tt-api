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
	functionUpdatePerson = debug.NewFunction(pkg, "UpdatePerson")
)

// UpdatePerson method
func UpdatePerson(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionUpdatePerson
	DebugVerbose(f, requestID, "")

	userID, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	personID, err := GetIntegerFromRequest(f, requestID, "id", data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "personID: %d", personID)

	user := model.FullPerson{ID: userID}
	err = user.LoadPersonTx(db)
	if err != nil {
		message := fmt.Sprintf("Could not load person [%d]", userID)
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
	}

	if userID == personID {
		err = user.CanEditSelf()
		if err != nil {
			message := "Not allowed to edit self"
			DebugVerbose(f, requestID, message)
			ReplyForbidden(requestID, client, replyTopic, message)
		}
	} else {
		err = user.CanEditOtherPeople()
		if err != nil {
			message := "Not allowed to edit other people"
			DebugVerbose(f, requestID, message)
			ReplyForbidden(requestID, client, replyTopic, message)
		}
	}

	err = model.UpdatePersonFieldsTx(db, personID, *data)
	if err != nil {
		message := fmt.Sprintf("problem updating person fields: userID: %d", userID)
		DebugVerbose(f, requestID, message)
		ReplyForbidden(requestID, client, replyTopic, message)
	}

	err = publisher.UpdatePublications(db, client, cfg)
	if err != nil {
		message := err.Error()
		DebugVerbose(f, requestID, message)
		ReplyInternalServerError(requestID, client, replyTopic, message)
	}

	ReplyOK(requestID, client, replyTopic)
}
