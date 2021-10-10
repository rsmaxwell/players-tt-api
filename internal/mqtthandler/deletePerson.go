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
	functionDeletePerson = debug.NewFunction(pkg, "DeletePerson")
)

// DeletePerson method
func DeletePerson(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionDeleteCourt
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
		return
	}

	if userID == personID {
		err = user.CanEditSelf()
		if err != nil {
			message := "Not allowed to delete self"
			DebugVerbose(f, requestID, message)
			ReplyForbidden(requestID, client, replyTopic, message)
		}
	} else {
		err = user.CanEditOtherPeople()
		if err != nil {
			message := "Not allowed to delete other people"
			DebugVerbose(f, requestID, message)
			ReplyForbidden(requestID, client, replyTopic, message)
		}
	}

	p := model.FullPerson{ID: personID}
	err = p.DeletePersonTx(db)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	ReplyOK(requestID, client, replyTopic)
}
