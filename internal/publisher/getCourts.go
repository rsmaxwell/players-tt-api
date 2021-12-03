package publisher

import (
	"database/sql"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	functionGetCourts = debug.NewFunction(pkg, "GetCourts")
)

// GetCourts method
func GetCourts(db *sql.DB, client mqtt.Client, cfg *config.Config) ([]Entry, error) {
	f := functionGetCourts
	f.DebugVerbose("")

	listOfCourts, err := model.ListCourtsTx(db)
	if err != nil {
		f.DebugVerbose(err.Error())
		return nil, err
	}

	entry := Entry{topic: "getCourts", object: listOfCourts}
	array := []Entry{entry}

	for _, court := range listOfCourts {
		court := model.Court{ID: court.ID}
		err := court.LoadCourtTx(db)
		if err != nil {
			f.DebugVerbose(err.Error())
			return nil, err
		}

		topic := fmt.Sprintf("getCourt/%d", court.ID)
		entry = Entry{topic: topic, object: court}
		array = append(array, entry)
	}

	return array, nil
}
