package main

import (
	"fmt"
	"os"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/model"

	_ "github.com/jackc/pgx/stdlib"
)

var (
	pkg          = debug.NewPackage("main")
	functionMain = debug.NewFunction(pkg, "main")
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

	f.Infof("Players CreateDB: Version: %s", basic.Version())

	// Read configuration
	cfg, err := config.Open(args.Configfile)
	if err != nil {
		message := "Error setting up"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	db, err := model.Connect(cfg)
	if err != nil {
		f.Errorf("Error Connecting to the database up")
		os.Exit(1)
	}
	defer db.Close()

	// Disconnect all users from the database, except ourselves
	sqlStatement := `
	SELECT pg_terminate_backend(pg_stat_activity.pid)
	FROM pg_stat_activity
	WHERE pg_stat_activity.datname = 'players'
	  AND pid <> pg_backend_pid()`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not disconnect other users from database"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	// Drop the database
	sqlStatement = `DROP DATABASE players`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := "Could not drop database"
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	fmt.Println("Successfully deleted the database.")
}
