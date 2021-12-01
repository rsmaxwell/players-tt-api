package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/mqtthandler"
	"github.com/rsmaxwell/players-tt-api/internal/publisher"
	"github.com/rsmaxwell/players-tt-api/internal/utils"

	"github.com/rsmaxwell/players-tt-api/model"
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

	args, err := cmdline.GetArguments()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	f.DebugInfo("Version: %s", basic.Version())

	if args.Version {
		fmt.Printf("Version: %s\n", basic.Version())
		fmt.Printf("BuildDate: %s\n", basic.BuildDate())
		fmt.Printf("GitCommit: %s\n", basic.GitCommit())
		fmt.Printf("GitBranch: %s\n", basic.GitBranch())
		fmt.Printf("GitURL: %s\n", basic.GitURL())
		os.Exit(0)
	}

	configfile := path.Join(args.Configdir, config.DefaultConfigFile)
	cfg, err = config.Open(configfile)
	if err != nil {
		f.Errorf("Error opening config: %s", configfile)
		os.Exit(1)
	}

	db, err = model.Connect(cfg)
	if err != nil {
		f.Errorf("Error Connecting to the database up")
		os.Exit(1)
	}
	defer db.Close()

	people, err := model.ListPeopleTx(db, "")
	if err != nil {
		f.Errorf("Could not list the people")
		os.Exit(1)
	}

	f.DebugInfo("People:")
	for _, person := range people {
		f.DebugInfo("    ID: %d, Knownas: %s, Email: %s", person.ID, person.Knownas, person.Email)
	}

	var host = cfg.Mqtt.Host
	var port = cfg.Mqtt.Port
	var broker = fmt.Sprintf("tcp://%s:%d", host, port)
	f.DebugVerbose("Broker: %s", broker)

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
		f.DumpError(err, "Could not connect to the mqtt broker on %s", broker)
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
