package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/mqtthandler"
	"github.com/rsmaxwell/players-tt-api/internal/publisher"
	"github.com/rsmaxwell/players-tt-api/internal/utils"
)

var (
	pkg                        = debug.NewPackage("main")
	functionMessagePubHandler  = debug.NewFunction(pkg, "messagePubHandler")
	functionConnectHandler     = debug.NewFunction(pkg, "connectHandler")
	functionConnectLostHandler = debug.NewFunction(pkg, "connectLostHandler")
	functionMain               = debug.NewFunction(pkg, "main")
	functionOnMessage          = debug.NewFunction(pkg, "onMessage")
)

const maxDuration time.Duration = 1<<63 - 1

var (
	handlers = map[string]mqtthandler.Handler{
		"register":     mqtthandler.Register,
		"signin":       mqtthandler.Signin,
		"getCourts":    mqtthandler.GetCourts,
		"getPeople":    mqtthandler.GetPeople,
		"getPerson":    mqtthandler.GetPerson,
		"updatePerson": mqtthandler.UpdatePerson,
		"getWaiters":   mqtthandler.GetWaiters,
		"refreshToken": mqtthandler.RefreshToken,
		"getCourt":     mqtthandler.GetCourt,
		"updateCourt":  mqtthandler.UpdateCourt,
		"createCourt":  mqtthandler.CreateCourt,
		"deleteCourt":  mqtthandler.DeleteCourt,
		"deletePerson": mqtthandler.DeletePerson,
		"fillCourt":    mqtthandler.FillCourt,
		"clearCourt":   mqtthandler.ClearCourt,
	}
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	f := functionMessagePubHandler
	f.DebugVerbose("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	f := functionConnectHandler
	f.DebugVerbose("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	f := functionConnectLostHandler
	f.DebugVerbose("Connect lost: %v\n", err)
	f.DebugVerbose("Exiting...\n")
	os.Exit(1)
}

var db *sql.DB
var cfg *config.Config
var requestID = 0

func main() {
	f := functionMain
	f.Verbosef("Read configuration and connect to the database")

	var err error
	db, cfg, err = config.Setup()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}
	defer db.Close()

	var host = cfg.Mqtt.Host
	var port = cfg.Mqtt.Port
	var broker = fmt.Sprintf("tcp://%s:%d", host, port)
	f.DebugVerbose("Broker: %s", broker)
	f.DebugVerbose("BuildID: %s", utils.BuildID())
	f.DebugVerbose("BuildDate: %s", utils.BuildDate())
	f.DebugVerbose("GitCommit: %s", utils.GitCommit())
	f.DebugVerbose("GitBranch: %s", utils.GitBranch())
	f.DebugVerbose("GitURL: %s", utils.GitURL())

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(cfg.Mqtt.ClientID)
	opts.SetUsername(cfg.Mqtt.Username)
	opts.SetPassword(cfg.Mqtt.Password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		f.DebugVerbose(err.Error())
		f.DumpError(err, "Could not connect")
		return
	}

	err = publisher.UpdatePublications(db, client, cfg)
	if err != nil {
		f.DebugVerbose(err.Error())
		f.DumpError(err, "Could not update publications")
		return
	}

	utils.Subscribe(client, "request", onMessage)
	time.Sleep(maxDuration)
}

var onMessage mqtt.MessageHandler = func(client mqtt.Client, message mqtt.Message) {
	f := functionOnMessage

	requestID = requestID + 1
	mqtthandler.DebugVerbose(f, requestID, "------------------------------------------------------------------------------------------------")

	payload := message.Payload()
	mqtthandler.DebugVerbose(f, requestID, "payload: %s", payload)

	var request mqtthandler.Request
	err := json.Unmarshal(payload, &request)
	if err != nil {
		mqtthandler.DebugVerbose(f, requestID, "Unexpected request: %s", payload)
		return
	}

	replyTopic, err := mqtthandler.GetReplyTopic(requestID, request)
	if err != nil {
		mqtthandler.DebugVerbose(f, requestID, "Unexpected 'replyTopic' in request: %s", err.Error())
		return
	}

	command, err := mqtthandler.GetCommand(requestID, request)
	if err != nil {
		message := fmt.Sprintf("Unexpected 'command' in request: %s", err.Error())
		mqtthandler.DebugVerbose(f, requestID, message)
		mqtthandler.ReplyBadRequest(requestID, client, replyTopic, message)
		return
	}

	handler := handlers[command]
	if handler == nil {
		message := fmt.Sprintf("Command not found: %s", command)
		mqtthandler.DebugVerbose(f, requestID, message)
		mqtthandler.ReplyBadRequest(requestID, client, replyTopic, message)
		return
	}

	data, err := mqtthandler.GetData(request)
	if err != nil {
		f.DebugVerbose("Unexpected 'data' in request: %s", err.Error())
		mqtthandler.ReplyBadRequest(requestID, client, replyTopic, err.Error())
		return
	}

	handler(db, cfg, requestID, client, replyTopic, data)
}
