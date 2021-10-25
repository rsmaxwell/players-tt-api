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

	// Read configuration and connect to the database
	cfg, err := config.Open(args.Configfile)
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}

	db, err := model.Connect(cfg)
	if err != nil {
		f.Errorf("Error Connecting to the database up")
		os.Exit(1)
	}
	defer db.Close()

	// Create the database
	sqlStatement := fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DatabaseName)
	_, err = db.Exec(sqlStatement)
	if err != nil {
		message := fmt.Sprintf("Could not create database: %s", cfg.Database.DatabaseName)
		f.Errorf(message)
		f.DumpSQLError(err, message, sqlStatement)
		os.Exit(1)
	}

	fmt.Printf("Successfully created database: %s\n", cfg.Database.DatabaseName)
}
