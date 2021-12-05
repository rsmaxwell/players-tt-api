package mqtthandler

import (
	"context"
	"database/sql"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionSignin = debug.NewFunction(pkg, "Signin")
)

func Signin(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionSignin
	DebugVerbose(f, requestID, "")

	signin, err := model.NewSigninFromMap(data)
	if err != nil {
		ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	email := signin.Username

	p, err := model.FindPersonByEmail(context.Background(), db, email)
	if err != nil {
		f.DebugVerbose("FindPersonByEmail returned err: %s", err.Error())
		ReplyBadRequest(requestID, client, replyTopic, "Not Authenticated")
		return
	}

	err = p.Authenticate(db, signin.Password)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "accessTokenExpiry:  %10s     expires at: %s", cfg.AccessTokenExpiry, time.Now().Add(cfg.AccessTokenExpiry).Round(time.Second))
	accessToken, err := basic.GenerateToken(p.ID, requestID, cfg.AccessTokenExpiry)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	DebugVerbose(f, requestID, "refreshTokenExpiry: %10s     expires at: %s", cfg.RefreshTokenExpiry, time.Now().Add(cfg.RefreshTokenExpiry).Round(time.Second))
	refreshToken, err := basic.GenerateToken(p.ID, requestID, cfg.RefreshTokenExpiry)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	var refreshDelta = int(cfg.ClientRefreshDelta / time.Second)
	DebugVerbose(f, requestID, "refreshDelta: %d", refreshDelta)

	reply := struct {
		Status       int    `json:"status"`
		Message      string `json:"message"`
		ID           int    `json:"id"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		RefreshDelta int    `json:"refreshDelta"`
	}{
		Status:       StatusOK,
		Message:      "ok",
		ID:           p.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		RefreshDelta: refreshDelta,
	}

	Reply(requestID, client, replyTopic, reply)
}
