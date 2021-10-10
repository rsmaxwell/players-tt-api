package mqtthandler

import (
	"database/sql"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionRegister = debug.NewFunction(pkg, "Register")
)

func Register(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionRegister
	DebugVerbose(f, requestID, "")

	registration, err := model.NewRegistrationFromMap(data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	p, err := registration.ToPerson()
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	err = p.SavePersonTx(db)
	if err != nil {
		pgx, ok := err.(pgx.PgError)
		if ok {
			err = codeerror.NewBadRequest(pgx.Message)
		}

		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	ReplyOK(requestID, client, replyTopic)
}
