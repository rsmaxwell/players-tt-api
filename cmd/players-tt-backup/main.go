package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/backup"
	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"

	"database/sql"

	_ "github.com/jackc/pgx/stdlib"
)

var (
	pkg                = debug.NewPackage("main")
	functionMain       = debug.NewFunction(pkg, "main")
	functionGetPeople  = debug.NewFunction(pkg, "getPeople")
	functionGetCourts  = debug.NewFunction(pkg, "getCourts")
	functionGetPlays   = debug.NewFunction(pkg, "getPlays")
	functionGetWaiters = debug.NewFunction(pkg, "GetWaiters")
)

func init() {
	debug.InitDump("com.rsmaxwell.players", "players-createdb", "https://server.rsmaxwell.co.uk/archiva")
}

// http://go-database-sql.org/retrieving.html
func main() {
	f := functionMain

	args, err := cmdline.GetArguments()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	if args.Version {
		fmt.Printf("Version: %s\n", basic.Version())
		fmt.Printf("BuildDate: %s\n", basic.BuildDate())
		fmt.Printf("GitCommit: %s\n", basic.GitCommit())
		fmt.Printf("GitBranch: %s\n", basic.GitBranch())
		fmt.Printf("GitURL: %s\n", basic.GitURL())
		os.Exit(0)
	}

	ctx := context.Background()

	f.Infof("Players backup: Version: %s", basic.Version())

	// Read configuration and connect to the database
	db, c, err := config.Setup(args.Configfile)
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}
	defer db.Close()

	var myBackup backup.Backup

	err = getPeople(ctx, db, &myBackup)
	if err != nil {
		message := "Could not get the people"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	err = getCourts(ctx, db, &myBackup)
	if err != nil {
		message := "Could not get the courts"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	err = getPlays(ctx, db, &myBackup)
	if err != nil {
		message := "Could not get the plays"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	err = getWaiters(ctx, db, &myBackup)
	if err != nil {
		message := "Could not get the waiters"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	// Marshal and write the backup to file
	bytearray, err := json.Marshal(&myBackup)
	if err != nil {
		message := "Could not Marshal backup"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	backupFile := filepath.Join(debug.RootDir(), "backup", "players.json")
	err = ioutil.WriteFile(backupFile, bytearray, 0644)
	if err != nil {
		message := "could not read backupFile file"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	fmt.Printf("Successfully populated the database: %s\n", c.Database.DatabaseName)
}

func getPeople(ctx context.Context, db *sql.DB, myBackup *backup.Backup) error {
	f := functionGetPeople

	// Query all the people in the person table
	sqlStatement := "SELECT * FROM " + model.PersonTable

	rows, err := db.Query(sqlStatement)
	if err != nil {
		message := "Could not select people"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	myBackup.PersonFieldsArray = []backup.PersonFields{}

	for rows.Next() {
		var p model.NullPerson
		err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Knownas, &p.Email, &p.Phone, &p.Hash, &p.Status)
		if err != nil {
			f.Errorf("Error: %t %v\n", err, err)
			return err
		}

		fields := make(map[string]interface{})
		fields["id"] = p.ID

		if p.FirstName.Valid {
			fields["firstname"] = p.FirstName.String
		}

		if p.LastName.Valid {
			fields["lastname"] = p.LastName.String
		}

		if p.Knownas.Valid {
			fields["displayname"] = p.Knownas.String
		}

		if p.Email.Valid {
			fields["email"] = p.Email.String
		}

		if p.Phone.Valid {
			fields["phone"] = p.Phone.String
		}

		if p.Hash.Valid {
			fields["hash"] = p.Hash.String
		}

		if p.Status.Valid {
			fields["status"] = p.Status.String
		}

		myBackup.PersonFieldsArray = append(myBackup.PersonFieldsArray, fields)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all the people"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func getCourts(ctx context.Context, db *sql.DB, myBackup *backup.Backup) error {
	f := functionGetCourts

	// Query all the courts in the courts table
	sqlStatement := "SELECT * FROM " + model.CourtTable

	rows, err := db.Query(sqlStatement)
	if err != nil {
		message := "Could not select from the court table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	myBackup.CourtFieldsArray = []backup.CourtFields{}

	for rows.Next() {
		var c model.NullCourt
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			f.Errorf("Error: %t %v\n", err, err)
			return err
		}

		court := make(map[string]interface{})
		court["id"] = c.ID

		if c.Name.Valid {
			court["name"] = c.Name.String
		}

		myBackup.CourtFieldsArray = append(myBackup.CourtFieldsArray, court)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all the courts"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func getPlays(ctx context.Context, db *sql.DB, myBackup *backup.Backup) error {
	f := functionGetPlays

	// Query all the plays in the playing table
	sqlStatement := "SELECT * FROM playing"

	rows, err := db.Query(sqlStatement)
	if err != nil {
		message := "Could not select plays"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	myBackup.Playing = []backup.Play{}

	var (
		person int
		court  int
	)
	for rows.Next() {
		err := rows.Scan(&person, &court)
		if err != nil {
			message := "Could not scan the play"
			f.Errorf(message)
			f.DumpError(err, message)
			return err
		}

		var play backup.Play
		play.Person = person
		play.Court = court

		myBackup.Playing = append(myBackup.Playing, play)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all the courts"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func getWaiters(ctx context.Context, db *sql.DB, myBackup *backup.Backup) error {
	f := functionGetWaiters

	// Query all the waiters in the waiting table
	sqlStatement := "SELECT * FROM waiting"

	rows, err := db.Query(sqlStatement)
	if err != nil {
		message := "Could not select waiters"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	myBackup.Waiting = []backup.Waiter{}

	var nw backup.NullWaiter
	for rows.Next() {
		err := rows.Scan(&nw.Person, &nw.Start)
		if err != nil {
			message := "Could not scan the waiter"
			f.Errorf(message)
			f.DumpError(err, message)
			return err
		}

		var w backup.Waiter
		w.Person = nw.Person
		w.Start = time.Now()

		if nw.Start.Valid {
			w.Start = nw.Start.Time
		}

		myBackup.Waiting = append(myBackup.Waiting, w)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all the courts"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}
