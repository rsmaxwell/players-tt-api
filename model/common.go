package model

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionSetup              = debug.NewFunction(pkg, "Setup")
	functionDeleteAllRecordsTx = debug.NewFunction(pkg, "DeleteAllRecordsTx")
	functionDeleteAllRecords   = debug.NewFunction(pkg, "deleteAllRecords")
	functionFillCourtTx        = debug.NewFunction(pkg, "FillCourtTx")
	functionFillCourt          = debug.NewFunction(pkg, "fillCourt")
	functionClearCourtTx       = debug.NewFunction(pkg, "ClearCourtTx")
	functionClearCourt         = debug.NewFunction(pkg, "clearCourt")
	functionEndTransaction     = debug.NewFunction(pkg, "EndTransaction")
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

	args, err := cmdline.GetArguments()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	// Read configuration
	configfile := path.Join(args.Configdir, config.DefaultConfigFile)
	cfg, err := config.Open(configfile)
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
	err = DeleteAllRecords(db)
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

// DeleteAllRecords
func DeleteAllRecords(db *sql.DB) error {
	f := functionDeleteAllRecords
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}
	defer EndTransaction(ctx, tx, db, err)

	err = deleteAllRecordsTx(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAllRecords removes all the records in the database
func deleteAllRecordsTx(ctx context.Context, db *sql.DB) error {
	f := functionDeleteAllRecordsTx

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
func FillCourt(db *sql.DB, courtID int) ([]Position, error) {
	f := functionFillCourt
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}
	defer EndTransaction(ctx, tx, db, err)

	positions, err := fillCourtTx(ctx, db, courtID)
	if err != nil {
		return nil, err
	}

	return positions, nil
}

// FillCourt
func fillCourtTx(ctx context.Context, db *sql.DB, courtID int) ([]Position, error) {
	f := functionFillCourtTx

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

			if personID <= 0 {
				message := "no more waiters"
				f.Infof(message)
				break
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
		err = person.LoadPersonTx(ctx, db)
		if err != nil {
			message := "Could not load player"
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}

		personId := PersonId{ID: player.Person, Knownas: person.Knownas}
		position := Position{Index: player.Position, PersonId: personId}
		positions = append(positions, position)
	}

	return positions, nil
}

// ClearCourt
func ClearCourt(db *sql.DB, courtID int) error {
	f := functionClearCourt
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}
	defer EndTransaction(ctx, tx, db, err)

	err = clearCourtTx(ctx, db, courtID)
	if err != nil {
		message := "Problem clearing court"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

// ClearCourt
func clearCourtTx(ctx context.Context, db *sql.DB, courtID int) error {
	f := functionClearCourtTx

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
		err = person.LoadPersonTx(ctx, db)
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

func EndTransaction(ctx context.Context, tx *sql.Tx, db *sql.DB, err error) error {
	f := functionEndTransaction
	f.DebugVerbose("")

	p := recover()
	if p != nil {
		f.DebugVerbose("Rollback on panic")
		tx.Rollback()
		panic(p)
	} else if err != nil {
		f.DebugVerbose("Rollback on error")
		tx.Rollback()
	} else {

		count, err := CheckConistency(ctx, db, false)
		if err != nil {
			f.Errorf("Rollback on failed consistency check")
			tx.Rollback()
			return err
		}
		if count > 0 {
			message := fmt.Sprintf("Rollback on inconsistant data: count: %d", count)
			f.Errorf(message)
			err = fmt.Errorf(message)
			tx.Rollback()
			return err
		}

		f.DebugVerbose("Commit on success")
		return tx.Commit()
	}

	return err
}
