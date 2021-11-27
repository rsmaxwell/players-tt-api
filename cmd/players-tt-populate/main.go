package main

import (
	"fmt"
	"os"

	"github.com/rsmaxwell/players-tt-api/internal/cmdline"
	"github.com/rsmaxwell/players-tt-api/model"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"

	_ "github.com/jackc/pgx/stdlib"
)

var (
	pkg                = debug.NewPackage("main")
	functionMain       = debug.NewFunction(pkg, "main")
	functionMakePeople = debug.NewFunction(pkg, "makePeople")
	functionMakeCourts = debug.NewFunction(pkg, "makeCourts")
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

	f.Infof("Players Populate: Version: %s", basic.Version())

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

	_, err = model.MakePeople(db)
	if err != nil {
		f.Errorf("Error making people")
		os.Exit(1)
	}

	/* courtIDs */
	_, err = model.MakeCourts(db)
	if err != nil {
		f.Errorf("Error making courts")
		os.Exit(1)
	}

	count, err := model.CheckConistencyTx(db, true)
	if err != nil {
		f.Errorf("Error checking consistency")
		os.Exit(1)
	}

	fmt.Printf("Made %d database updates", count)
}
