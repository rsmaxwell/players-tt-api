package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

// Waiter type
type Waiter struct {
	Person int       `json:"person"`
	Start  time.Time `json:"start"`
}

// NullWaiter type
type NullWaiter struct {
	Person int
	Start  sql.NullTime
}

type DisplayWaiter struct {
	PersonID int    `json:"personID"`
	Knownas  string `json:"knownas"`
	Start    int64  `json:"start"`
}

const (
	// WaitingTable is the name of the waiting table
	WaitingTable = "waiting"
	WaitingIndex = "first_waiting"
)

var (
	functionListWaitersTx        = debug.NewFunction(pkg, "ListWaitersTx")
	functionListWaiters          = debug.NewFunction(pkg, "ListWaiters")
	functionListWaitersForPerson = debug.NewFunction(pkg, "ListWaitersForPerson")
	functionGetFirstWaiter       = debug.NewFunction(pkg, "GetFirstWaiter")
	functionRemoveWaiter         = debug.NewFunction(pkg, "RemoveWaiter")
)

// ListWaiters returns the list of waiters
func ListWaitersTx(db *sql.DB) ([]Waiter, error) {
	f := functionListWaitersTx
	ctx := context.Background()

	// Create a new context, and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}

	listOfWaiters, err := ListWaiters(ctx, db)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
		return nil, err
	}

	return listOfWaiters, nil
}

// ListWaiters returns the list of waiters
func ListWaiters(ctx context.Context, db *sql.DB) ([]Waiter, error) {
	f := functionListWaiters

	sqlStatement := "SELECT * FROM " + WaitingTable + " ORDER BY start ASC"

	rows, err := db.QueryContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not get list the waiters"
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Waiter
	for rows.Next() {

		var nw NullWaiter
		err := rows.Scan(&nw.Person, &nw.Start)
		if err != nil {
			message := "Could not scan the waiter"
			f.DumpError(err, message)
			return nil, err
		}

		var w Waiter
		w.Person = nw.Person

		if nw.Start.Valid {
			w.Start = nw.Start.Time
		}

		list = append(list, w)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the waiters"
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// ListWaitersForPerson returns the list of waiters for a person
func ListWaitersForPerson(ctx context.Context, db *sql.DB, id int) ([]Waiter, error) {
	f := functionListWaitersForPerson

	fields := "person, start"
	sqlStatement := "SELECT " + fields + " FROM " + WaitingTable + " WHERE person=$1"

	rows, err := db.Query(sqlStatement, id)
	if err != nil {
		message := "Could not get list the waiters"
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Waiter
	for rows.Next() {

		var nw NullWaiter
		err := rows.Scan(&nw.Person, &nw.Start)
		if err != nil {
			message := "Could not scan the waiter"
			f.DumpError(err, message)
			return nil, err
		}

		var w Waiter
		w.Person = nw.Person

		if nw.Start.Valid {
			w.Start = nw.Start.Time
		}

		list = append(list, w)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the waiters"
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// Get first GetFirstWaiter
func GetFirstWaiter(ctx context.Context, db *sql.DB) (int, error) {
	f := functionGetFirstWaiter

	fields := "person"
	sqlStatement := "SELECT " + fields + " FROM " + WaitingTable + " ORDER BY start LIMIT 1"
	rows, err := db.QueryContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not get the first waiter"
		f.DumpSQLError(err, message, sqlStatement)
		return 0, err
	}
	defer rows.Close()

	count := 0
	var id int
	for rows.Next() {
		count++
		err := rows.Scan(&id)
		if err != nil {
			message := "Could not scan the first player"
			f.DumpError(err, message)
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		message := "Error get the first player"
		f.DumpError(err, message)
		return 0, err
	}
	if count < 1 {
		message := "There were no waiters"
		f.DumpError(err, message)
		return 0, err
	}

	return id, nil
}

func AddWaiter(ctx context.Context, db *sql.DB, personID int) error {
	f := functionRemoveWaiter

	start := time.Now()

	fields := "person, start"
	values := "$1, $2"
	sqlStatement := "INSERT INTO " + WaitingTable + " (" + fields + ") VALUES (" + values + ")"

	_, err := db.ExecContext(ctx, sqlStatement, personID, start)
	if err != nil {
		message := "Could not insert into " + WaitingTable
		d := f.DumpSQLError(err, message, sqlStatement)
		data := struct {
			PersonID int
			Start    time.Time
		}{
			PersonID: personID,
			Start:    start,
		}
		bytes, _ := json.MarshalIndent(data, "", "    ")
		d.AddByteArray("values.json", bytes)
		return err
	}

	return nil
}

func RemoveWaiter(ctx context.Context, db *sql.DB, personID int) error {
	f := functionRemoveWaiter

	sqlStatement := "DELETE FROM " + WaitingTable + " WHERE person=$1"
	rows, err := db.QueryContext(ctx, sqlStatement, personID)
	if err != nil {
		message := "Could not delete the waiter"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	return nil
}
