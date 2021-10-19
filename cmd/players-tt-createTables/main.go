package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/model"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"

	_ "github.com/jackc/pgx/stdlib"
)

var (
	pkg                 = debug.NewPackage("main")
	functionMain        = debug.NewFunction(pkg, "main")
	functionDropTable   = debug.NewFunction(pkg, "dropTable")
	functionTableExists = debug.NewFunction(pkg, "tableExists")
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

	f.Infof("Players CreateTables: Version: %s", basic.Version())

	AdminFirstName, ok := os.LookupEnv("PLAYERS_ADMIN_FIRST_NAME")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_FIRST_NAME not set")
		os.Exit(1)
	}

	AdminLastName, ok := os.LookupEnv("PLAYERS_ADMIN_LAST_NAME")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_LAST_NAME not set")
		os.Exit(1)
	}

	AdminKnownas, ok := os.LookupEnv("PLAYERS_ADMIN_KNOWNAS")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_KNOWNAS not set")
		os.Exit(1)
	}

	AdminEmail, ok := os.LookupEnv("PLAYERS_ADMIN_EMAIL")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_EMAIL not set")
		os.Exit(1)
	}

	AdminPhone, ok := os.LookupEnv("PLAYERS_ADMIN_PHONE")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_PHONE not set")
		os.Exit(1)
	}

	AdminPassword, ok := os.LookupEnv("PLAYERS_ADMIN_PASSWORD")
	if !ok {
		f.Errorf("PLAYERS_ADMIN_PASSWORD not set")
		os.Exit(1)
	}

	db, c, err := config.Setup(args.Configfile)
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}
	defer db.Close()

	// Drop the tables
	err = dropTable(ctx, db, model.PlayingTable)
	if err != nil {
		return
	}

	err = dropTable(ctx, db, model.WaitingTable)
	if err != nil {
		return
	}

	err = dropTable(ctx, db, model.PersonTable)
	if err != nil {
		return
	}

	err = dropTable(ctx, db, model.CourtTable)
	if err != nil {
		return
	}

	// Create the people table
	sqlStatement := `
		CREATE TABLE ` + model.PersonTable + ` (
			id SERIAL PRIMARY KEY,
			firstname VARCHAR(255) NOT NULL,
			lastname VARCHAR(255) NOT NULL,
			knownas VARCHAR(32) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			phone VARCHAR(32) NOT NULL UNIQUE,
			hash VARCHAR(255) NOT NULL,	
			status VARCHAR(32) NOT NULL
		 )`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create person table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the person_email index
	sqlStatement = "CREATE INDEX person_email ON " + model.PersonTable + " ( email )"
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create person_email index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the court table
	sqlStatement = `
		CREATE TABLE ` + model.CourtTable + ` (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255)
		 )`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create court table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the playing table
	sqlStatement = `
		CREATE TABLE ` + model.PlayingTable + ` (
			court    INT NOT NULL,
			person   INT NOT NULL,
			position INT NOT NULL,		

			PRIMARY KEY (court, person, position),

			CONSTRAINT person FOREIGN KEY(person) REFERENCES person(id),
			CONSTRAINT court FOREIGN KEY(court)  REFERENCES court(id)
		 )`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create playing table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the playing_court index
	sqlStatement = "CREATE INDEX playing_court ON " + model.PlayingTable + " ( court )"
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the playing_person index
	sqlStatement = "CREATE INDEX playing_person ON " + model.PlayingTable + " ( person )"
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create playing_person index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the waiting table
	sqlStatement = `
		CREATE TABLE ` + model.WaitingTable + ` (
			person INT PRIMARY KEY,
			start  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT person FOREIGN KEY(person) REFERENCES person(id)
		 )`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create waiting table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Create the waiting index
	sqlStatement = "CREATE INDEX " + model.WaitingIndex + " ON " + model.WaitingTable + " ( start )"
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	fmt.Printf("Successfully created Tables in the database: %s\n", c.Database.DatabaseName)

	peopleData := []model.Registration{
		{FirstName: AdminFirstName, LastName: AdminLastName, Knownas: AdminKnownas, Email: AdminEmail, Phone: AdminPhone, Password: AdminPassword},
	}

	peopleIDs := make(map[int]int)
	for i, r := range peopleData {

		p, err := r.ToPerson()
		if err != nil {
			message := "Could not register person"
			f.Errorf(message)
			f.DumpError(err, message)
			os.Exit(1)
		}

		p.Status = model.StatusAdmin

		err = p.SavePersonTx(db)
		if err != nil {
			message := fmt.Sprintf("Could not save person: firstName: %s, lastname: %s, email: %s", p.FirstName, p.LastName, p.Email)
			f.Errorf(message)
			f.DumpError(err, message)
			os.Exit(1)
		}

		peopleIDs[i] = p.ID

		fmt.Printf("Added person:\n")
		fmt.Printf("    FirstName: %s\n", p.FirstName)
		fmt.Printf("    LastName:  %s\n", p.LastName)
		fmt.Printf("    Knownas:   %s\n", p.Knownas)
		fmt.Printf("    Email:     %s\n", p.Email)
		fmt.Printf("    Password:  %s\n", r.Password)
		fmt.Printf("    Hash:      %s\n", p.Hash)
		fmt.Printf("    Status:    %s\n", p.Status)
	}
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	f := functionTableExists

	sqlStatement := fmt.Sprintf("SELECT EXISTS ( SELECT FROM %s WHERE schemaname = '%s' AND tablename = '%s' )", "pg_tables", "public", table)
	row := db.QueryRow(sqlStatement)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		message := fmt.Sprintf("Could not drop table: %s", table)
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return false, err
	}

	return exists, nil
}

func dropTable(ctx context.Context, db *sql.DB, table string) error {
	f := functionDropTable

	exists, err := tableExists(ctx, db, table)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	sqlStatement := `DROP TABLE ` + table
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := fmt.Sprintf("Could not drop table: %s", table)
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	for i := 0; (i < 10) && exists; i++ {
		exists, err = tableExists(ctx, db, table)
		if err != nil {
			return err
		}

		if exists {
			time.Sleep(1 * time.Second)
		}
	}

	if exists {
		message := fmt.Sprintf("Could not drop table: %s", table)
		f.Errorf(message)
		err = errors.New(message)
	}

	return err
}
