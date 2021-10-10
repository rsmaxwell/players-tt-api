package model

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

// Player type
type Player struct {
	Person   int `json:"person"`
	Court    int `json:"court"`
	Position int `json:"position"`
}

const (
	// PlayingTable is the name of the table
	PlayingTable = "playing"
)

var (
	functionListPlayers          = debug.NewFunction(pkg, "ListPlayers")
	functionAddPlayer            = debug.NewFunction(pkg, "AddPlayer")
	functionRemovePlayer         = debug.NewFunction(pkg, "RemovePlayer")
	functionListPlayersForPerson = debug.NewFunction(pkg, "ListPlayersForPerson")
	functionListPlayersForCourt  = debug.NewFunction(pkg, "ListPlayersForCourt")
)

// AddPlayer
func AddPlayer(ctx context.Context, db *sql.DB, personID int, courtID int, position int) error {
	f := functionAddPlayer

	fields := "person, court, position"
	values := "$1, $2, $3"
	sqlStatement := "INSERT INTO " + PlayingTable + " (" + fields + ") VALUES (" + values + ")"

	_, err := db.ExecContext(ctx, sqlStatement, personID, courtID, position)
	if err != nil {
		message := "Could not insert into " + PlayingTable
		f.Errorf(message)
		d := f.DumpSQLError(err, message, sqlStatement)
		d.AddObject("player", struct {
			personID int
			courtID  int
			position int
		}{
			personID: personID,
			courtID:  courtID,
			position: position,
		})
		return err
	}

	return nil
}

// RemovePlayer
func RemovePlayer(ctx context.Context, db *sql.DB, personID int) error {
	f := functionRemovePlayer

	sqlStatement := "DELETE FROM " + PlayingTable + " WHERE person=$1"

	_, err := db.ExecContext(ctx, sqlStatement, personID)
	if err != nil {
		message := "Could not delete row from " + PlayingTable
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return nil
}

// ListPlayers
func ListPlayers(ctx context.Context, db *sql.DB) ([]Player, error) {
	f := functionListPlayers

	fields := "court, person, position"
	sqlStatement := "SELECT " + fields + " FROM " + PlayingTable

	rows, err := db.Query(sqlStatement)
	if err != nil {
		message := "Could not list the players"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Player
	for rows.Next() {

		var player Player
		err := rows.Scan(&player.Person, &player.Court, &player.Position)
		if err != nil {
			message := "Could not scan the player"
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}

		list = append(list, player)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the players"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// ListPlayersForPerson
func ListPlayersForPerson(ctx context.Context, db *sql.DB, personID int) ([]Player, error) {
	f := functionListPlayersForPerson

	fields := "court, person, position"
	sqlStatement := "SELECT " + fields + " FROM " + PlayingTable + " WHERE person=$1"

	rows, err := db.QueryContext(ctx, sqlStatement, personID)
	if err != nil {
		message := "Could not get list the players"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Player
	for rows.Next() {

		var player Player
		err := rows.Scan(&player.Person, &player.Court, &player.Position)
		if err != nil {
			message := "Could not scan the player"
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}

		list = append(list, player)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the players"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// ListPlayersForCourt
func ListPlayersForCourt(ctx context.Context, db *sql.DB, courtID int) ([]Player, error) {
	f := functionListPlayersForCourt

	fields := "person, court, position"
	sqlStatement := "SELECT " + fields + " FROM " + PlayingTable + " WHERE court=$1"

	rows, err := db.QueryContext(ctx, sqlStatement, courtID)
	if err != nil {
		message := "Could not get list the players"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Player
	for rows.Next() {

		var player Player
		err := rows.Scan(&player.Person, &player.Court, &player.Position)
		if err != nil {
			message := "Could not scan the player"
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}

		list = append(list, player)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the players"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// Dump writes the player to a dump file
func (p *Player) Dump(d *debug.Dump) {

	bytearray, err := json.Marshal(p)
	if err != nil {
		return
	}

	title := "player.json"
	d.AddByteArray(title, bytearray)
}
