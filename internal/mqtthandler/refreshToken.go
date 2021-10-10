package mqtthandler

import (
	"database/sql"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionRefreshToken = debug.NewFunction(pkg, "RefreshToken")
)

// RefreshToken method
func RefreshToken(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionRefreshToken
	DebugVerbose(f, requestID, "")

	// *********************************************************************
	// * Check the existing access token is valid
	// *********************************************************************
	_, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	// *********************************************************************
	// * Validate the refreshToken
	// *********************************************************************
	refreshToken, err := GetStringFromRequest(f, requestID, "refreshToken", data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	claims, err := basic.ValidateToken(refreshToken)
	if err != nil {
		message := fmt.Sprintf("refreshToken not valid: %s", err.Error())
		DebugVerbose(f, requestID, message)
		ReplyUnAuthorised(requestID, client, replyTopic, message)
		return
	}

	// *********************************************************************
	// * Create a new access token
	// *********************************************************************
	DebugVerbose(f, requestID, "accessTokenExpiry:  %10s     expires at: %s", cfg.AccessTokenExpiry, time.Now().Add(cfg.AccessTokenExpiry))
	newAccessToken, err := basic.GenerateToken(claims.ID, claims.Request, cfg.AccessTokenExpiry)
	if err != nil {
		ReplyInternalServerError(requestID, client, replyTopic, err.Error())
		return
	}

	reply := struct {
		Status      int    `json:"status"`
		Message     string `json:"message"`
		AccessToken string `json:"accessToken"`
	}{
		Status:      StatusOK,
		Message:     "ok",
		AccessToken: newAccessToken,
	}

	Reply(requestID, client, replyTopic, reply)
}
