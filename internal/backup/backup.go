package backup

import (
	"database/sql"
	"time"
)

// Backup type
type Backup struct {
	PersonFieldsArray []PersonFields `json:"people"`
	CourtFieldsArray  []CourtFields  `json:"courts"`
	Playing           []Play         `json:"playing"`
	Waiting           []Waiter       `json:"waiting"`
}

// PersonFields type
type PersonFields map[string]interface{}

// CourtFields type
type CourtFields map[string]interface{}

// Play type
type Play struct {
	Person int `json:"person"`
	Court  int `json:"court"`
}

// NullWaiter type
type NullWaiter struct {
	Person int
	Start  sql.NullTime
}

// Waiter type
type Waiter struct {
	Person int       `json:"person"`
	Start  time.Time `json:"start"`
}

// Indexes type
type Indexes struct {
	People map[int]int
	Courts map[int]int
}

// NewIndexes is a constructor
func NewIndexes() *Indexes {
	i := new(Indexes)
	i.People = make(map[int]int)
	i.Courts = make(map[int]int)
	return i
}
