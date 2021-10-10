package utils

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	pkg                       = debug.NewPackage("utils")
	functionGetStringFromMap  = debug.NewFunction(pkg, "GetStringFromMap")
	functionGetIntegerFromMap = debug.NewFunction(pkg, "GetIntegerFromMap")
	functionPublishObject     = debug.NewFunction(pkg, "PublishObject")
	functionPublish           = debug.NewFunction(pkg, "Publish")
	functionSubscribe         = debug.NewFunction(pkg, "Subscribe")
)

func GetStringFromMap(key string, data *map[string]interface{}) (string, error) {
	f := functionGetStringFromMap

	object, ok := (*data)[key]
	if !ok {
		return "", fmt.Errorf("could not find the key [%s]", key)
	}
	value, ok := object.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	f.DebugVerbose("key: %s, value: %s", key, value)
	return value, nil
}

func GetIntegerFromMap(key string, data *map[string]interface{}) (int, error) {
	f := functionGetIntegerFromMap

	object, ok := (*data)[key]
	if !ok {
		return 0, fmt.Errorf("could not find the key [%s]", key)
	}

	value, ok := object.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	f.DebugVerbose("key: %s, value: %d", key, int(value))
	return int(value), nil
}

func PublishObject(client mqtt.Client, topic string, object interface{}) (mqtt.Token, error) {
	f := functionPublishObject

	b, err := json.Marshal(object)
	if err != nil {
		f.DebugVerbose(err.Error())
		return &mqtt.DummyToken{}, err
	}

	return Publish(client, topic, string(b)), nil
}

func Publish(client mqtt.Client, topic string, message string) mqtt.Token {
	f := functionPublish

	f.DebugVerbose("Publish to [%s] message: %s", topic, message)
	var qos byte = 1
	var retained bool = true
	return client.Publish(topic, qos, retained, message)
}

func Subscribe(client mqtt.Client, topic string, callback mqtt.MessageHandler) error {
	f := functionSubscribe

	var qos byte = 1
	if token := client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		f.DumpError(token.Error(), "Could not subscribe to topic [%s]", topic)
	}

	f.DebugVerbose("Subscribed to topic [%s]", topic)
	return nil
}
