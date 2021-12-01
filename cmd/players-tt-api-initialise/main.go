package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"

	"github.com/rsmaxwell/players-tt-api/model"
)

var (
	pkg                            = debug.NewPackage("main")
	functionMain                   = debug.NewFunction(pkg, "main")
	functionConnect                = debug.NewFunction(pkg, "connect")
	functionCreateDatabase         = debug.NewFunction(pkg, "createDatabase")
	functionCreateTablesInDatabase = debug.NewFunction(pkg, "createTablesInDatabase")
	functionInitialiseDatabaseTx   = debug.NewFunction(pkg, "initialiseDatabaseTx")
	functionInitialiseDatabase     = debug.NewFunction(pkg, "initialiseDatabase")
	functionDropTables             = debug.NewFunction(pkg, "dropTables")
	functionDropTable              = debug.NewFunction(pkg, "dropTable")
	functionTableExists            = debug.NewFunction(pkg, "tableExists")
	functionCreateTables           = debug.NewFunction(pkg, "createTables")
	functionCreateAdminUser        = debug.NewFunction(pkg, "createAdminUser")
	functionDatabaseExists         = debug.NewFunction(pkg, "databaseExists")
)

var cfg *config.Config

func main() {
	f := functionMain

	args, err := cmdline.GetArguments()
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	f.DebugInfo("Version: %s", basic.Version())

	if args.Version {
		fmt.Printf("Version: %s\n", basic.Version())
		fmt.Printf("BuildDate: %s\n", basic.BuildDate())
		fmt.Printf("GitCommit: %s\n", basic.GitCommit())
		fmt.Printf("GitBranch: %s\n", basic.GitBranch())
		fmt.Printf("GitURL: %s\n", basic.GitURL())
		os.Exit(0)
	}

	configfile := path.Join(args.Configdir, config.DefaultConfigFile)
	cfg, err = config.Open(configfile)
	if err != nil {
		f.Errorf("Error opening config: %s", configfile)
		os.Exit(1)
	}

	err = createDatabase(cfg.Database.DatabaseName)
	if err != nil {
		f.Errorf("Error creating database: %s", cfg.Database.DatabaseName)
		os.Exit(1)
	}

	err = createTablesInDatabase()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func createDatabase(databaseName string) error {
	f := functionCreateDatabase
	f.DebugVerbose("")

	f.DebugInfo("Connect to postgres (without database)")
	connectionString := cfg.ConnectionStringBasic()
	db, err := connect(cfg, connectionString)
	if err != nil {
		f.Errorf("Error connecting to postgres: %s", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	f.DebugInfo("Check database exists")
	found, err := databaseExists(db, cfg.Database.DatabaseName)
	if err != nil {
		message := fmt.Sprintf("Problem checking the database '%s'", cfg.Database.DatabaseName)
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	if found {
		f.DebugInfo("Database already exists")
	} else {
		f.DebugInfo("Create the database: %s", cfg.Database.DatabaseName)
		sqlStatement := fmt.Sprintf("CREATE DATABASE %s", databaseName)
		_, err := db.Exec(sqlStatement)
		if err != nil {
			message := fmt.Sprintf("Could not create database: %s", databaseName)
			f.Errorf(message)
			f.DumpSQLError(err, message, sqlStatement)
			return err
		}
	}

	return nil
}

func createTablesInDatabase() error {
	f := functionCreateTablesInDatabase

	f.DebugInfo("Connect to postgres (with database)")
	connectionString := fmt.Sprintf("%s/%s", cfg.ConnectionStringBasic(), cfg.Database.DatabaseName)
	db, err := connect(cfg, connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	f.DebugInfo("Initialising database")
	err = initialiseDatabaseTx(db)
	if err != nil {
		return err
	}

	return nil
}

func connect(cfg *config.Config, connectionString string) (*sql.DB, error) {
	f := functionConnect

	f.DebugInfo("Connect to postgres")
	driverName := cfg.DriverName()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		message := fmt.Sprintf("Could not connect to '%s' with connectionString: '%s'", driverName, connectionString)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	return db, err
}

func initialiseDatabaseTx(db *sql.DB) error {
	f := functionInitialiseDatabaseTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = initialiseDatabase(ctx, db)
	if err != nil {
		tx.Rollback()
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
		return err
	}

	return nil
}

func initialiseDatabase(ctx context.Context, db *sql.DB) error {
	f := functionInitialiseDatabase
	f.DebugVerbose("")

	err := dropTables(ctx, db)
	if err != nil {
		return err
	}

	err = createTables(ctx, db)
	if err != nil {
		return err
	}

	err = createAdminUser(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	f := functionCreateTables
	f.DebugVerbose("")

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
	_, err := db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create person table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the person_email index
	sqlStatement = "CREATE INDEX person_email ON " + model.PersonTable + " ( email )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create person_email index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the court table
	sqlStatement = `
		CREATE TABLE ` + model.CourtTable + ` (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255)
		 )`
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create court table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
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
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the playing_court index
	sqlStatement = "CREATE INDEX playing_court ON " + model.PlayingTable + " ( court )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the playing_person index
	sqlStatement = "CREATE INDEX playing_person ON " + model.PlayingTable + " ( person )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_person index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the waiting table
	sqlStatement = `
		CREATE TABLE ` + model.WaitingTable + ` (
			person INT PRIMARY KEY,
			start  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT person FOREIGN KEY(person) REFERENCES person(id)
		 )`
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create waiting table"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the waiting index
	sqlStatement = "CREATE INDEX " + model.WaitingIndex + " ON " + model.WaitingTable + " ( start )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	f.DebugInfo("Successfully created Tables")
	return nil
}

func dropTables(ctx context.Context, db *sql.DB) error {
	f := functionDropTables
	f.DebugVerbose("")

	// Drop the tables
	err := dropTable(ctx, db, model.PlayingTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, model.WaitingTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, model.PersonTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, model.CourtTable)
	if err != nil {
		return err
	}

	return nil
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
	_, err = db.ExecContext(ctx, sqlStatement)
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

func createAdminUser(ctx context.Context, db *sql.DB) error {
	f := functionCreateAdminUser
	f.DebugVerbose("")

	AdminFirstName, ok := os.LookupEnv("PLAYERS_ADMIN_FIRST_NAME")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_FIRST_NAME not set")
	}

	AdminLastName, ok := os.LookupEnv("PLAYERS_ADMIN_LAST_NAME")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_LAST_NAME not set")
	}

	AdminKnownas, ok := os.LookupEnv("PLAYERS_ADMIN_KNOWNAS")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_KNOWNAS not set")
	}

	AdminEmail, ok := os.LookupEnv("PLAYERS_ADMIN_EMAIL")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_EMAIL not set")
	}

	AdminPhone, ok := os.LookupEnv("PLAYERS_ADMIN_PHONE")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_PHONE not set")
	}

	AdminPassword, ok := os.LookupEnv("PLAYERS_ADMIN_PASSWORD")
	if !ok {
		return fmt.Errorf("PLAYERS_ADMIN_PASSWORD not set")
	}

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

		err = p.SavePerson(ctx, db)
		if err != nil {
			message := fmt.Sprintf("Could not save person: firstName: %s, lastname: %s, email: %s", p.FirstName, p.LastName, p.Email)
			f.Errorf(message)
			f.DumpError(err, message)
			os.Exit(1)
		}

		peopleIDs[i] = p.ID

		f.DebugInfo("Added person:")
		f.DebugInfo("    ID:        %d", p.ID)
		f.DebugInfo("    FirstName: %s", p.FirstName)
		f.DebugInfo("    LastName:  %s", p.LastName)
		f.DebugInfo("    Knownas:   %s", p.Knownas)
		f.DebugInfo("    Email:     %s", p.Email)
		f.DebugInfo("    Password:  %s", r.Password)
		f.DebugInfo("    Hash:      %s", p.Hash)
		f.DebugInfo("    Status:    %s", p.Status)
	}

	return nil
}

func databaseExists(db *sql.DB, databaseName string) (bool, error) {
	f := functionDatabaseExists
	f.DebugVerbose("")

	sqlStatement := "SELECT COUNT(*) FROM pg_catalog.pg_database WHERE datname = '" + databaseName + "'"
	row := db.QueryRow(sqlStatement)

	var count int
	err := row.Scan(&count)
	if err != nil {
		message := "problem checking the database exists"
		f.Errorf(message + ": " + err.Error())
		f.DumpError(err, message)
		return false, err
	}

	return (count > 0), nil
}
