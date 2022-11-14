package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionUpdateGame           = debug.NewFunction(pkg, "UpdateGame")
	functionUpdateGameTx         = debug.NewFunction(pkg, "UpdateGameTx")
	functionCheckOriginalPlayer  = debug.NewFunction(pkg, "checkOriginalPlayer")
	functionCheckNewPlayer       = debug.NewFunction(pkg, "checkNewPlayer")
	functionCheckPlayerIsWaiting = debug.NewFunction(pkg, "checkPlayerIsWaiting")
)

type GamePosition struct {
	Value    *PersonId
	Original *PersonId
}

type GameData struct {
	Court     int
	Positions map[int]*GamePosition
}

// UpdateGame
func UpdateGame(db *sql.DB, gameData *GameData) error {
	f := functionUpdateGame
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}
	defer EndTransaction(ctx, tx, db, err)

	err = updateGameTx(ctx, db, gameData)
	if err != nil {
		return err
	}

	count, err := CheckConistency(ctx, db, false)
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

// UpdateGame
func updateGameTx(ctx context.Context, db *sql.DB, gameData *GameData) error {
	f := functionUpdateGameTx

	players, err := GetPlayersForCourtAsMap(ctx, db, gameData.Court)
	if err != nil {
		message := "Could not list players"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	if len(gameData.Positions) > NumberOfCourtPositions {
		message := fmt.Sprintf("Unexpected number of game positions: game: %d, #positions: %d", gameData.Court, len(gameData.Positions))
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	if len(players) > NumberOfCourtPositions {
		message := fmt.Sprintf("Unexpected number of players on court: game: %d, #players: %d", gameData.Court, len(players))
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	for index, position := range gameData.Positions {

		var personId *int = nil
		player, ok := players[index]
		if ok {
			personId = &player.Person
			f.DebugVerbose("personId: %d", *personId)
		}

		err := checkOriginalPlayer(index, personId, position.Original)
		if err != nil {
			return err
		}

		changed, err := checkNewPlayer(index, personId, position.Value)
		if err != nil {
			return err
		}
		if !changed {
			continue
		}

		if personId != nil {
			err = MakePlayerWaitTx(ctx, db, *personId)
			if err != nil {
				message := fmt.Sprintf("Could not make player [%d] on court [%d], wait", personId, gameData.Court)
				f.Errorf(message)
				f.DumpError(err, message)
				return err
			}
		}
	}

	waiters, err := ListWaitersTx(ctx, db)
	if err != nil {
		message := "Could not list waiters"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	for index, position := range gameData.Positions {

		var personId *int = nil
		player, ok := players[index]
		if ok {
			personId = &player.Person
			f.DebugVerbose("personId: %d", *personId)
		}

		err := checkOriginalPlayer(index, personId, position.Original)
		if err != nil {
			return err
		}

		changed, err := checkNewPlayer(index, personId, position.Value)
		if err != nil {
			return err
		}
		if !changed {
			continue
		}

		if position.Value != nil {
			err = checkPlayerIsWaiting(position.Value, waiters)
			if err != nil {
				return err
			}
			err = MakePlayerPlayTx(ctx, db, position.Value.ID, gameData.Court, index)
			if err != nil {
				message := fmt.Sprintf("Could not make player [%d] on court [%d], wait", position.Value.ID, gameData.Court)
				f.Errorf(message)
				f.DumpError(err, message)
				return err
			}
		}
	}

	return nil
}

// checkOriginalPlayer
func checkOriginalPlayer(index int, personId *int, originalPlayerId *PersonId) error {
	f := functionCheckOriginalPlayer

	if personId == nil {
		if originalPlayerId != nil {
			message := fmt.Sprintf("the player at position [%d] has changed. [%d] --> (nil)", index, originalPlayerId.ID)
			f.Errorf(message)
			return fmt.Errorf(message)
		}
	} else if originalPlayerId == nil {
		message := fmt.Sprintf("the player at position [%d] has changed. (nil) --> [%d]", index, *personId)
		f.Errorf(message)
		return fmt.Errorf(message)
	} else if *personId != originalPlayerId.ID {
		message := fmt.Sprintf("the player at position [%d] has changed. [%d] --> [%d]", index, originalPlayerId.ID, *personId)
		f.Errorf(message)
		return fmt.Errorf(message)
	}

	return nil
}

// checkNewPlayer
func checkNewPlayer(index int, personId *int, newPlayerId *PersonId) (bool, error) {
	f := functionCheckNewPlayer

	if personId == nil {
		if newPlayerId == nil {
			f.Infof("the player at position [%d] is unchanged. (nil) --> (nil)", index)
			return false, nil
		}
	} else if newPlayerId == nil {
	} else if *personId == newPlayerId.ID {
		f.Infof("the player at position [%d] is unchanged. [%d] --> [%d]", index, newPlayerId.ID, *personId)
		return false, nil
	}

	return true, nil
}

// checkPlayerIsWaiting
func checkPlayerIsWaiting(playerId *PersonId, waiters []Waiter) error {
	f := functionCheckPlayerIsWaiting

	found := false
	for _, waiter := range waiters {
		if playerId.ID == waiter.Person {
			found = true
			break
		}
	}

	if !found {
		message := fmt.Sprintf("cannot make player [%d] play as the player is not waiting", playerId.ID)
		f.Infof(message)
		return fmt.Errorf(message)
	}

	return nil
}
