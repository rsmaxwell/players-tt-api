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
	functionGetPeople = debug.NewFunction(pkg, "GetPeople")

	filters = make(map[string]string)
)

func init() {
	filters["all"] = ``
	filters["players"] = `WHERE status = 'player'`
	filters["inactive"] = `WHERE status = 'inactive'`
	filters["suspended"] = `WHERE status = 'suspended'`
}

// GetPeople method
func GetPeople(db *sql.DB, client mqtt.Client, cfg *config.Config) ([]Entry, error) {
	f := functionGetPeople
	f.DebugVerbose("")

	array := []Entry{}

	for filterName, whereClause := range filters {

		listOfFullPeople, err := model.ListPeopleTx(db, whereClause)
		if err != nil {
			f.DebugVerbose("filterName: %s, error: %s", filterName, err.Error())
			return nil, err
		}

		listOfPeople := []model.Person{}
		for _, fullPerson := range listOfFullPeople {
			person := *fullPerson.ToLimited()
			listOfPeople = append(listOfPeople, person)

			topic := fmt.Sprintf("getPerson/%d", person.ID)
			entry := Entry{topic: topic, object: listOfPeople}
			array = append(array, entry)
		}

		topic := fmt.Sprintf("getPeople/%s", filterName)
		entry := Entry{topic: topic, object: listOfPeople}
		array = append(array, entry)
	}

	return array, nil
}
