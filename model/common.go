package model

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionSetup            = debug.NewFunction(pkg, "Setup")
	functionDeleteAllRecords = debug.NewFunction(pkg, "DeleteAllRecords")
	functionFillCourtTx      = debug.NewFunction(pkg, "FillCourtTx")
	functionFillCourt        = debug.NewFunction(pkg, "FillCourt")
	functionClearCourtTx     = debug.NewFunction(pkg, "ClearCourtTx")
	functionClearCourt       = debug.NewFunction(pkg, "ClearCourt")
)

var (
	// MetricsData containing metrics
	MetricsData Metrics
)

// Metrics structure
type Metrics struct {
	StatusCodes map[int]int `json:"statusCodes"`
}

func init() {
	MetricsData = Metrics{}
	MetricsData.StatusCodes = make(map[int]int)
}

// Setup function
func Setup(t *testing.T) (func(t *testing.T), *sql.DB, *config.Config) {
	f := functionSetup
	ctx := context.Background()

	args, err := cmdline.GetArguments()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	// Read configuration
	cfg, err := config.Open(args.Configfile)
	if err != nil {
		f.Errorf("Error setting up")
		t.FailNow()
	}

	db, err := Connect(cfg)
	if err != nil {
		f.Errorf("Error Connecting to the database up")
		os.Exit(1)
	}
	defer db.Close()

	// Delete all the records
	err = DeleteAllRecords(ctx, db)
	if err != nil {
		f.Errorf("Error delete all the records")
		t.FailNow()
	}

	// Populate
	err = Populate(db)
	if err != nil {
		f.Errorf("Could not populate the database")
		t.FailNow()
	}

	return func(t *testing.T) {
		db.Close()
	}, db, cfg
}

// DeleteAllRecords removes all the records in the database
func DeleteAllRecords(ctx context.Context, db *sql.DB) error {
	f := functionDeleteAllRecords

	sqlStatement := "DELETE FROM " + PlayingTable
	_, err := db.Exec(sqlStatement)
	if err != nil {
		message := "Could not delete all from playing"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	sqlStatement = "DELETE FROM " + WaitingTable
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not delete all from waiting"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	sqlStatement = "DELETE FROM " + CourtTable
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not delete all from courts"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	sqlStatement = "DELETE FROM " + PersonTable + " WHERE status != '" + StatusAdmin + "'"
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not delete all from people"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return nil
}

// FillCourt
func FillCourtTx(db *sql.DB, courtID int) ([]Position, error) {
	f := functionFillCourtTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}

	positions, err := FillCourt(ctx, db, courtID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
	}

	count, err := CheckConistencyTx(db, false)
	if err != nil {
		f.Errorf("Error checking consistency")
		return nil, err
	}
	if count > 0 {
		message := fmt.Sprintf("Inconsistant data: count: %d", count)
		f.Errorf(message)
		err = fmt.Errorf(message)
		return nil, err
	}

	return positions, nil
}

// FillCourt
func FillCourt(ctx context.Context, db *sql.DB, courtID int) ([]Position, error) {
	f := functionFillCourt

	players, err := ListPlayersForCourt(ctx, db, courtID)
	if err != nil {
		message := "Could not list players"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	mapOfPlayers := make(map[int]*Player)
	for _, player := range players {
		p := player                        // take a copy of each object ...
		mapOfPlayers[player.Position] = &p // ... so their references are actually different!
	}

	changes := 0
	positions := make([]Position, 0)
	for index := 0; index < NumberOfCourtPositions; index++ {

		var ok bool
		var player *Player
		var personID int

		if player, ok = mapOfPlayers[index]; !ok {
			changes++

			personID, err = GetFirstWaiter(ctx, db)
			if err != nil {
				message := "Could not get the first waiter"
				f.Errorf(message)
				f.DumpError(err, message)
				return nil, err
			}

			err = RemoveWaiter(ctx, db, personID)
			if err != nil {
				message := "Could not remove the waiter"
				f.Errorf(message)
				f.DumpError(err, message)
				return nil, err
			}

			err = AddPlayer(ctx, db, personID, courtID, index)
			if err != nil {
				message := "Could not add player"
				f.Errorf(message)
				f.DumpError(err, message)
				return nil, err
			}
			p := Player{Person: personID, Court: courtID, Position: index}
			player = &p
		}

		person := FullPerson{ID: player.Person}
		err = person.LoadPerson(ctx, db)
		if err != nil {
			message := "Could not load player"
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}

		var position = Position{Index: player.Position, PersonID: player.Person, DisplayName: person.Knownas}
		positions = append(positions, position)
	}

	return positions, nil
}

// ClearCourt
func ClearCourtTx(db *sql.DB, courtID int) error {
	f := functionClearCourtTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = ClearCourt(ctx, db, courtID)
	if err != nil {
		tx.Rollback()
		message := "Problem clearing court"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	count, err := CheckConistencyTx(db, false)
	if err != nil {
		f.Errorf("Error checking consistency")
		return err
	}
	if count > 0 {
		message := fmt.Sprintf("Inconsistant data: count: %d", count)
		f.Errorf(message)
		err = fmt.Errorf(message)
		return err
	}

	return nil
}

// ClearCourt
func ClearCourt(ctx context.Context, db *sql.DB, courtID int) error {
	f := functionClearCourt

	players, err := ListPlayersForCourt(ctx, db, courtID)
	if err != nil {
		message := "Could not list players"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	for _, player := range players {
		err = RemovePlayer(ctx, db, player.Person)
		if err != nil {
			message := "Could not remove player"
			f.Errorf(message)
			f.DumpError(err, message)
			return err
		}

		person := FullPerson{ID: player.Person}
		err = person.LoadPerson(ctx, db)
		if err != nil {
			message := "Could not load player"
			f.Errorf(message)
			f.DumpError(err, message)
			return err
		}

		if person.Status == StatusPlayer {
			err = AddWaiter(ctx, db, player.Person)
			if err != nil {
				message := "Could not add waiter"
				f.Errorf(message)
				f.DumpError(err, message)
				return err
			}
		}
	}

	return nil
}

// EqualIntArray tells whether a and b contain the same elements NOT in-order order
func EqualIntArray(x, y []int) bool {

	if x == nil {
		return y == nil
	} else if y == nil {
		return false
	}

	if len(x) != len(y) {
		return false
	}

	xMap := make(map[int]int)
	yMap := make(map[int]int)

	for _, xElem := range x {
		xMap[xElem]++
	}
	for _, yElem := range y {
		yMap[yElem]++
	}

	for xMapKey, xMapVal := range xMap {
		if yMap[xMapKey] != xMapVal {
			return false
		}
	}
	return true
}
