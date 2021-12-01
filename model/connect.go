package model

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionConnect       = debug.NewFunction(pkg, "Connect")
	functionDatabaseCheck = debug.NewFunction(pkg, "DatabaseCheck")
)

const (
	Invalid_Catalog_Name = "3D000"
	Undefined_Table      = "42P01"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	f := functionConnect
	f.DebugVerbose("")

	driverName := cfg.DriverName()
	connectionString := cfg.ConnectionString()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		message := fmt.Sprintf("Could not connect to postgres: driverName: %s, connectionString:%s", driverName, connectionString)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}

	return db, err
}

func DatabaseCheck(db *sql.DB) (bool, error) {
	f := functionDatabaseCheck
	f.DebugVerbose("")

	var count int
	sqlStatement := "SELECT COUNT(*) FROM " + CourtTable
	err := db.QueryRow(sqlStatement).Scan(&count)

	ok := true
	if err != nil {
		if err2, ok2 := err.(pgx.PgError); ok2 {
			if err2.Code == Invalid_Catalog_Name {
				f.DebugInfo("%s: PgError.Code: %s  (Invalid Catalog Name)", err.Error(), err2.Code)
				return false, nil
			} else if err2.Code == Undefined_Table {
				f.DebugInfo("%s: PgError.Code: %s  (Undefined Table)", err.Error(), err2.Code)
				return false, nil
			} else {
				message := fmt.Sprintf("%s: PgError.Code: %s", err.Error(), err2.Code)
				f.Errorf(message)
				f.DumpError(err, message)
				return false, err
			}
		}
	}

	return ok, err
}
