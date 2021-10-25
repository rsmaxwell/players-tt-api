package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionConnect              = debug.NewFunction(pkg, "Connect")
	functionInitialiseDatabaseTx = debug.NewFunction(pkg, "initialiseDatabaseTx")
	functionInitialiseDatabase   = debug.NewFunction(pkg, "initialiseDatabase")
	functionCreateTables         = debug.NewFunction(pkg, "createTables")
	functionDropTables           = debug.NewFunction(pkg, "dropTables")
	functionDropTable            = debug.NewFunction(pkg, "dropTable")
	functionTableExists          = debug.NewFunction(pkg, "tableExists")
	functionDatabaseExists       = debug.NewFunction(pkg, "databaseExists")
	functionCreateDatabase       = debug.NewFunction(pkg, "createDatabase")
	functionDatabaseCheck        = debug.NewFunction(pkg, "databaseCheck")
)

const (
	Invalid_Catalog_Name = "3D000"
	Undefined_Table      = "42P01"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	f := functionConnect

	// Connect to the database & database
	driverName := cfg.DriverName()
	connectionString := cfg.ConnectionString()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		message := fmt.Sprintf("Could not connect to the database: driverName: %s, connectionString:%s", driverName, connectionString)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	ok, err := databaseCheck(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if ok {
		return db, nil
	}

	db.Close()

	// Connect to the database (no database)
	driverName = cfg.DriverName()
	connectionString = cfg.ConnectionStringBasic()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err = sql.Open(driverName, connectionString)
	if err != nil {
		message := fmt.Sprintf("Could not connect to the database: driverName: %s, connectionString:%s", driverName, connectionString)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	ok, err = databaseExists(db, cfg.Database.DatabaseName)
	if err != nil {
		message := fmt.Sprintf("Problem checking the database '%s' exists", cfg.Database.DatabaseName)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}
	if !ok {
		err = createDatabase(db, cfg.Database.DatabaseName)
		if err != nil {
			message := fmt.Sprintf("Problem checking the database '%s' exists", cfg.Database.DatabaseName)
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}
	}

	db.Close()

	// Reconnect to the database & database
	driverName = cfg.DriverName()
	connectionString = cfg.ConnectionString()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err = sql.Open(driverName, connectionString)
	if err != nil {
		message := fmt.Sprintf("Could not connect to the database: driverName: %s, connectionString:%s", driverName, connectionString)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	ok, err = databaseCheck(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if !ok {
		err = initialiseDatabaseTx(db)
		if err != nil {
			message := fmt.Sprintf("Problem initialising the database '%s'", cfg.Database.DatabaseName)
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}
	}

	ok, err = databaseCheck(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if !ok {
		db.Close()
		return nil, fmt.Errorf("problem initialising database")
	}

	return db, nil
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
		return nil
	}

	err = createTables(ctx, db)
	if err != nil {
		return nil
	}

	return nil
}

func dropTables(ctx context.Context, db *sql.DB) error {
	f := functionDropTables
	f.DebugVerbose("")

	// Drop the tables
	err := dropTable(ctx, db, PlayingTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, WaitingTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, PersonTable)
	if err != nil {
		return err
	}

	err = dropTable(ctx, db, CourtTable)
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
		CREATE TABLE ` + PersonTable + ` (
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
	sqlStatement = "CREATE INDEX person_email ON " + PersonTable + " ( email )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create person_email index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the court table
	sqlStatement = `
		CREATE TABLE ` + CourtTable + ` (
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
		CREATE TABLE ` + PlayingTable + ` (
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
	sqlStatement = "CREATE INDEX playing_court ON " + PlayingTable + " ( court )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the playing_person index
	sqlStatement = "CREATE INDEX playing_person ON " + PlayingTable + " ( person )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_person index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Create the waiting table
	sqlStatement = `
		CREATE TABLE ` + WaitingTable + ` (
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
	sqlStatement = "CREATE INDEX " + WaitingIndex + " ON " + WaitingTable + " ( start )"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not create playing_court index"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	fmt.Printf("Successfully created Tables\n")
	return nil
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

func databaseCheck(db *sql.DB) (bool, error) {
	f := functionDatabaseCheck
	f.DebugVerbose("")

	var count int
	sqlStatement := "SELECT COUNT(*) FROM " + CourtTable
	err := db.QueryRow(sqlStatement).Scan(&count)

	ok := true
	if err != nil {
		if err2, ok2 := err.(pgx.PgError); ok2 {
			if err2.Code == Invalid_Catalog_Name {
				ok = false
				err = nil
			} else if err2.Code == Undefined_Table {
				ok = false
				err = nil
			} else {
				ok = false
			}
		}
	}

	return ok, err
}

func createDatabase(db *sql.DB, databaseName string) error {
	f := functionCreateDatabase
	f.DebugVerbose("")

	sqlStatement := fmt.Sprintf("CREATE DATABASE %s", databaseName)
	_, err := db.Exec(sqlStatement)
	if err != nil {
		message := fmt.Sprintf("Could not create database: %s", databaseName)
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return nil
}
