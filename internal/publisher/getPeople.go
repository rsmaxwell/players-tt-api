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
	functionGetPeople       = debug.NewFunction(pkg, "GetPeople")
	functionGetPeopleAll    = debug.NewFunction(pkg, "GetPeopleAll")
	functionGetPeopleFilter = debug.NewFunction(pkg, "GetPeopleFilter")

	filters = make(map[string]string)
)

func init() {
	filters["players"] = `WHERE status = 'player'`
	filters["inactive"] = `WHERE status = 'inactive'`
	filters["suspended"] = `WHERE status = 'suspended'`
}

// GetPeople method
func GetPeople(db *sql.DB, client mqtt.Client, cfg *config.Config) ([]Entry, error) {
	f := functionGetPeople
	f.DebugVerbose("")

	array := []Entry{}

	items, err := getPeopleAll(db, client, cfg)
	if err != nil {
		return nil, err
	}
	array = append(array, items...)

	for filterName, whereClause := range filters {
		items, err = getPeopleFilter(db, client, cfg, filterName, whereClause)
		if err != nil {
			return nil, err
		}
		array = append(array, items...)
	}

	return array, nil
}

func getPeopleAll(db *sql.DB, client mqtt.Client, cfg *config.Config) ([]Entry, error) {
	f := functionGetPeopleAll
	f.DebugVerbose("")

	filterName := "all"
	whereClause := ""
	array := []Entry{}

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
		entry := Entry{topic: topic, object: person}
		array = append(array, entry)
	}

	topic := fmt.Sprintf("getPeople/%s", filterName)
	entry := Entry{topic: topic, object: listOfPeople}
	array = append(array, entry)

	return array, nil
}

func getPeopleFilter(db *sql.DB, client mqtt.Client, cfg *config.Config, filterName string, whereClause string) ([]Entry, error) {
	f := functionGetPeopleFilter
	f.DebugVerbose("filterName: [%s], whereClause: [%s]", filterName, whereClause)

	array := []Entry{}

	listOfFullPeople, err := model.ListPeopleTx(db, whereClause)
	if err != nil {
		f.DebugVerbose("filterName: %s, error: %s", filterName, err.Error())
		return nil, err
	}

	listOfPeople := []model.Person{}
	for _, fullPerson := range listOfFullPeople {
		person := *fullPerson.ToLimited()
		listOfPeople = append(listOfPeople, person)
	}

	topic := fmt.Sprintf("getPeople/%s", filterName)
	entry := Entry{topic: topic, object: listOfPeople}
	array = append(array, entry)

	return array, nil
}
