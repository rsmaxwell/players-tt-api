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
	functionUpdateGame        = debug.NewFunction(pkg, "UpdateGame")
	functionParseGameData     = debug.NewFunction(pkg, "parseGameData")
	functionParseGamePosition = debug.NewFunction(pkg, "parseGamePosition")
)

// UpdateGame method
func UpdateGame(db *sql.DB, cfg *config.Config, requestID int, client mqtt.Client, replyTopic string, data *map[string]interface{}) {
	f := functionUpdateGame
	DebugVerbose(f, requestID, "")

	userID, err := checkAuthenticated(requestID, data)
	if err != nil {
		ReplyUnAuthorised(requestID, client, replyTopic, err.Error())
		return
	}

	user := model.FullPerson{ID: userID}
	err = user.LoadPerson(db)
	if err != nil {
		message := fmt.Sprintf("Could not load person [%d]", userID)
		DebugVerbose(f, requestID, err.Error())
		ReplyInternalServerError(requestID, client, replyTopic, message)
		return
	}

	err = user.CanEditGame()
	if err != nil {
		message := fmt.Sprintf("Person [%d] is not allowed to edit game", userID)
		DebugVerbose(f, requestID, err.Error())
		ReplyForbidden(requestID, client, replyTopic, message)
		return
	}

	gameData, err := parseGameData(requestID, data)
	if err != nil {
		message := "Problem parsing request data"
		DebugVerbose(f, requestID, err.Error())
		ReplyForbidden(requestID, client, replyTopic, message)
		return
	}

	err = model.UpdateGame(db, gameData)
	if err != nil {
		message := "problem updating Game"
		DebugVerbose(f, requestID, err.Error())
		ReplyForbidden(requestID, client, replyTopic, message)
		return
	}

	err = publisher.UpdatePublications(db, client, cfg)
	if err != nil {
		message := "problem updating publications"
		DebugVerbose(f, requestID, err.Error())
		ReplyInternalServerError(requestID, client, replyTopic, message)
	}

	ReplyOK(requestID, client, replyTopic)
}

// Parse the request data
func parseGameData(requestID int, data *map[string]interface{}) (*model.GameData, error) {
	f := functionParseGameData
	DebugVerbose(f, requestID, "")

	gameData := new(model.GameData)

	// *** "court" *********

	x, found := (*data)["court"]
	if !found {
		return nil, fmt.Errorf("missing field: 'court'")
	}

	id1, ok := x.(int)
	if ok {
		gameData.Court = id1
	} else {
		id2, ok := x.(float64)
		if ok {
			gameData.Court = int(id2)
		} else {
			return nil, fmt.Errorf("unexpected type: 'court': %T", x)
		}
	}

	// *** "positions" *********

	y, found := (*data)["positions"]
	if !found {
		return nil, fmt.Errorf("missing field: 'positions'")
	}

	array, ok := y.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type: 'positions': %T", y)
	}

	gameData.Positions = make(map[int]*model.GamePosition)

	for _, item := range array {

		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type: 'item': %T", item)
		}

		for key, value := range itemMap {
			DebugVerbose(f, requestID, "key: %s, value: %#v", key, value)
		}

		index, err := parseGamePositionIndex(requestID, &itemMap, 0, model.NumberOfCourtPositions)
		if err != nil {
			return nil, err
		}

		value, err := parseGamePosition(requestID, &itemMap, "value")
		if err != nil {
			return nil, err
		}

		original, err := parseGamePosition(requestID, &itemMap, "original")
		if err != nil {
			return nil, err
		}

		previous := gameData.Positions[index]
		if previous != nil {
			return nil, fmt.Errorf("repeated index in request data: index: %d", index)
		}

		position := model.GamePosition{Value: value, Original: original}
		gameData.Positions[index] = &position
	}

	return gameData, nil
}

// Parse the request data to get the index field
func parseGamePositionIndex(requestID int, data *map[string]interface{}, min int, max int) (int, error) {
	f := functionParseGamePosition
	DebugVerbose(f, requestID, "")

	x, found := (*data)["index"]
	if !found {
		return 0, fmt.Errorf("missing field: '%s'", "index")
	}

	index, ok := x.(int)
	if !ok {
		z, ok := x.(float64)
		if !ok {
			return 0, fmt.Errorf("unexpected type: 'index': %T", x)
		}
		index = int(z)
	}

	if index < min {
		return 0, fmt.Errorf("unexpected request data: index: %d", index)
	}

	if index >= max {
		return 0, fmt.Errorf("unexpected request data: index: %d", index)
	}

	return index, nil
}

// Parse the request data
func parseGamePosition(requestID int, data *map[string]interface{}, fieldName string) (*model.PersonId, error) {
	f := functionParseGamePosition
	DebugVerbose(f, requestID, "")

	x, found := (*data)[fieldName]
	if !found {
		return nil, nil
	}

	myMap, ok := x.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type: '%s': %T", fieldName, x)
	}

	y, found := myMap["id"]
	if !found {
		return nil, fmt.Errorf("missing field: 'id'")
	}

	id, ok := y.(int)
	if !ok {
		z, ok := y.(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type: 'id': %T", y)
		}
		id = int(z)
	}

	z, found := myMap["knownas"]
	if !found {
		return nil, fmt.Errorf("missing field: 'knownas'")
	}

	knownas, ok := z.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: 'knownas': %T", y)
	}

	personId := model.PersonId{ID: id, Knownas: knownas}
	return &personId, nil
}
