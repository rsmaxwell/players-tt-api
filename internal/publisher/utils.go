package publisher

import (
	"database/sql"
	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/utils"
)

var (
	pkg = debug.NewPackage("publisher")

	functionUpdatePublications = debug.NewFunction(pkg, "UpdatePublications")
)

type Entry struct {
	topic  string
	object interface{}
}

type Handler func(*sql.DB, mqtt.Client, *config.Config) ([]Entry, error)

var (
	handlers = []Handler{
		GetCourts,
		GetPeople,
		GetWaiters,
	}

	previous = map[string]string{}
)

// UpdatePublications method
func UpdatePublications(db *sql.DB, client mqtt.Client, cfg *config.Config) error {
	f := functionUpdatePublications
	f.DebugVerbose("")

	history := map[string]string{}

	for _, handler := range handlers {
		entries, err := handler(db, client, cfg)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			bytes, err := json.Marshal(entry.object)
			if err != nil {
				f.DebugVerbose(err.Error())
				return err
			}
			newMessage := string(bytes)
			history[entry.topic] = newMessage

			oldMessage, ok := previous[entry.topic]
			if ok && (oldMessage == newMessage) {
				continue
			}

			utils.Publish(client, entry.topic, newMessage)
		}
	}

	previous = history

	return nil
}
