package mqtthandler

import (
	"database/sql"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	pkg                        = debug.NewPackage("mqtthandler")
	functionGetStringField     = debug.NewFunction(pkg, "GetStringField")
	functionGetData            = debug.NewFunction(pkg, "GetData")
	functionPublish            = debug.NewFunction(pkg, "Publish")
	functionSubscribe          = debug.NewFunction(pkg, "Subscribe")
	functionCheckAuthenticated = debug.NewFunction(pkg, "checkAuthenticated")
)

type Request map[string]interface{}

type Handler func(*sql.DB, *config.Config, int, mqtt.Client, string, *map[string]interface{})

func GetCommand(requestID int, request Request) (string, error) {
	return GetStringField(requestID, request, "command")
}

func GetReplyTopic(requestID int, request Request) (string, error) {
	return GetStringField(requestID, request, "replyTopic")
}

func GetStringField(requestID int, request Request, field string) (string, error) {
	f := functionGetStringField

	var object interface{}
	var ok bool

	if object, ok = request[field]; !ok {
		message := fmt.Sprintf("the request did not contain the field '%s'", field)
		DebugVerbose(f, requestID, message)
		return "", fmt.Errorf(message)
	}
	replyTopic, ok := object.(string)
	if !ok {
		message := fmt.Sprintf("the field '%s' is not a string: %#v", field, object)
		f.DebugVerbose(message)
		return "", fmt.Errorf(message)
	}

	return replyTopic, nil
}

func GetData(request Request) (*map[string]interface{}, error) {
	f := functionGetData

	var object interface{}
	var ok bool

	field := "data"

	if object, ok = request[field]; !ok {
		message := fmt.Sprintf("the request did not contain the field '%s'", field)
		f.DebugVerbose(message)
		return nil, fmt.Errorf(message)
	}
	data, ok := object.(map[string]interface{})
	if !ok {
		message := fmt.Sprintf("the field '%s' is not of type 'Data': %#v", field, object)
		f.DebugVerbose(message)
		return nil, fmt.Errorf(message)
	}

	return &data, nil
}

func Publish(requestID int, client mqtt.Client, topic string, object interface{}) (mqtt.Token, error) {
	f := functionPublish
	// f.DebugVerbose("")

	b, err := json.Marshal(object)
	if err != nil {
		fmt.Println(err)
		return &mqtt.DummyToken{}, err
	}
	message := string(b)
	DebugVerbose(f, requestID, "Publish to [%s] message: %s", topic, message)
	var qos byte = 0
	var retained bool = false
	return client.Publish(topic, qos, retained, message), nil
}

func Subscribe(requestID int, client mqtt.Client, topic string, callback mqtt.MessageHandler) {
	f := functionSubscribe

	var qos byte = 1
	if token := client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	DebugVerbose(f, requestID, "Subscribed to topic [%s]", topic)
}

func PublishResponse(requestID int, client mqtt.Client, topic string, status int, message string) {

	reply := struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: message,
	}

	Publish(requestID, client, topic, reply)
}

func GetFormattedRequestID(requestID int) string {
	return fmt.Sprintf("[request:%d]", requestID)
}

func DebugInfo(f *debug.Function, requestID int, format string, a ...interface{}) {
	if f.Level() >= debug.VerboseLevel {
		requestID := GetFormattedRequestID(requestID)
		message := fmt.Sprintf(format, a...)
		f.DebugInfo("%s %s", requestID, message)
	}
}

func DebugError(f *debug.Function, requestID int, format string, a ...interface{}) {
	if f.Level() >= debug.VerboseLevel {
		requestID := GetFormattedRequestID(requestID)
		message := fmt.Sprintf(format, a...)
		f.DebugError("%s %s", requestID, message)
	}
}

func DebugVerbose(f *debug.Function, requestID int, format string, a ...interface{}) {
	if f.Level() >= debug.VerboseLevel {
		requestID := GetFormattedRequestID(requestID)
		message := fmt.Sprintf(format, a...)
		f.DebugVerbose("%s %s", requestID, message)
	}
}

const (
	StatusOK                  = 0
	StatusBadRequest          = 400
	StatusUnAuthorised        = 401
	StatusForbidden           = 403
	StatusInternalServerError = 500
)

func Reply(requestID int, client mqtt.Client, topic string, data interface{}) {
	Publish(requestID, client, topic, data)
}

func ReplyOK(requestID int, client mqtt.Client, topic string) {
	PublishResponse(requestID, client, topic, StatusOK, "ok")
}

func ReplyBadRequest(requestID int, client mqtt.Client, topic string, message string) {
	PublishResponse(requestID, client, topic, StatusBadRequest, message)
}

func ReplyForbidden(requestID int, client mqtt.Client, topic string, message string) {
	PublishResponse(requestID, client, topic, StatusForbidden, message)
}

func ReplyUnAuthorised(requestID int, client mqtt.Client, topic string, message string) {
	PublishResponse(requestID, client, topic, StatusUnAuthorised, message)
}

func ReplyInternalServerError(requestID int, client mqtt.Client, topic string, message string) {
	PublishResponse(requestID, client, topic, StatusInternalServerError, message)
}

// checkAuthenticated method
func checkAuthenticated(requestID int, data *map[string]interface{}) (int, error) {
	f := functionCheckAuthenticated

	accessToken, err := GetStringFromRequest(f, requestID, "accessToken", data)
	if err != nil {
		message := "missing accessToken"
		f.DebugError(message)
		return 0, fmt.Errorf(message)
	}

	claims, err := basic.ValidateToken(accessToken)
	if err != nil {
		return 0, err
	}

	DebugVerbose(f, requestID, fmt.Sprintf("jwtClaims: user:%d, request:%d", claims.ID, claims.Request))

	return claims.ID, nil
}

func GetStringFromRequest(f *debug.Function, requestID int, key string, data *map[string]interface{}) (string, error) {

	object, ok := (*data)[key]
	if !ok {
		return "", fmt.Errorf("could not find the key [%s]", key)
	}
	value, ok := object.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	DebugVerbose(f, requestID, "key: %s, value: %s", key, value)
	return value, nil
}

func GetIntegerFromRequest(f *debug.Function, requestID int, key string, data *map[string]interface{}) (int, error) {

	object, ok := (*data)[key]
	if !ok {
		return 0, fmt.Errorf("could not find the key [%s]", key)
	}

	value, ok := object.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	DebugVerbose(f, requestID, "key: %s, value: %d", key, int(value))
	return int(value), nil
}

func Dump(f *debug.Function, requestID int, format string, a ...interface{}) *debug.Dump {
	d := f.Dump(format, a...)
	d.AddString("RequestID", GetFormattedRequestID(requestID))
	return d
}

func DumpError(f *debug.Function, err error, requestID int, format string, a ...interface{}) *debug.Dump {
	d := f.DumpError(err, format, a...)
	d.AddString("RequestID", GetFormattedRequestID(requestID))
	return d
}
