package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

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
	pkg                   = debug.NewPackage("main")
	functionMain          = debug.NewFunction(pkg, "main")
	functionInsertPeople  = debug.NewFunction(pkg, "insertPeople")
	functionInsertCourts  = debug.NewFunction(pkg, "insertCourts")
	functionInsertPlays   = debug.NewFunction(pkg, "insertPlays")
	functionInsertWaiters = debug.NewFunction(pkg, "insertWaiters")
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

	f.Infof("Players CreateDB: Version: %s", basic.Version())

	// Read configuration and connect to the database
	db, c, err := config.Setup(args.Configfile)
	if err != nil {
		f.Errorf("Error setting up")
		os.Exit(1)
	}
	defer db.Close()

	// Read the the backup file
	backupFile := filepath.Join(debug.RootDir(), "backup", "players.json")

	bytearray, err := ioutil.ReadFile(backupFile)
	if err != nil {
		message := "could not read backupFile file"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	var myBackup backup.Backup
	err = json.Unmarshal(bytearray, &myBackup)
	if err != nil {
		message := "could not Unmarshal configuration"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	err = model.DeleteAllRecords(ctx, db)
	if err != nil {
		message := "could not delete record"
		f.Errorf(message)
		f.DumpError(err, message)
		os.Exit(1)
	}

	indexes := backup.NewIndexes()

	err = insertPeople(ctx, db, &myBackup, indexes)
	if err != nil {
		message := "could not insert people"
		f.Errorf(message)
		os.Exit(1)
	}

	err = insertCourts(ctx, db, &myBackup, indexes)
	if err != nil {
		message := "could not insert courts"
		f.Errorf(message)
		os.Exit(1)
	}

	err = insertPlays(ctx, db, &myBackup, indexes)
	if err != nil {
		message := "could not insert plays"
		f.Errorf(message)
		os.Exit(1)
	}

	err = insertWaiters(ctx, db, &myBackup, indexes)
	if err != nil {
		message := "could not insert plays"
		f.Errorf(message)
		os.Exit(1)
	}

	fmt.Printf("Successfully restored the database: %s\n", c.Database.DatabaseName)
}

func insertPeople(ctx context.Context, db *sql.DB, myBackup *backup.Backup, indexes *backup.Indexes) error {
	f := functionInsertPeople

	// Insert people into the people table

	for _, fieldsMap := range myBackup.PersonFieldsArray {
		fields := ""
		values := ""
		separator := ""
		id1 := 0

		if value, ok := fieldsMap["id"]; ok {
			if num, ok := value.(float64); ok {
				id1 = int(num)
			}
		}

		if value, ok := fieldsMap["firstname"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "firstname"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["lastname"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "lastname"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["displayname"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "displayname"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["username"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "username"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["email"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "email"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["phone"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "phone"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["hash"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "hash"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["status"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "status"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		if value, ok := fieldsMap["status"]; ok {
			if str, ok := value.(string); ok {
				if str == model.StatusAdmin {
					continue
				}
			}
		}

		sqlStatement := "INSERT INTO " + model.PersonTable + " (" + fields + ") VALUES	(" + values + ") RETURNING id"

		var id2 int
		err := db.QueryRowContext(ctx, sqlStatement).Scan(&id2)
		if err != nil {
			message := "Could not insert into people"
			f.Errorf(message)
			f.DumpSQLError(err, message, sqlStatement)
			return err
		}
		f.Infof("Inserted person: %d", id2)

		indexes.People[id1] = id2
	}

	return nil
}

func insertCourts(ctx context.Context, db *sql.DB, myBackup *backup.Backup, indexes *backup.Indexes) error {
	f := functionInsertCourts

	// Insert courts into the courts table

	for _, fieldsMap := range myBackup.CourtFieldsArray {
		fields := ""
		values := ""
		separator := ""
		id1 := 0

		if value, ok := fieldsMap["id"]; ok {
			if num, ok := value.(float64); ok {
				id1 = int(num)
			}
		}

		if value, ok := fieldsMap["name"]; ok {
			if str, ok := value.(string); ok {
				fields = fields + separator + "name"
				values = values + separator + basic.Quote(str)
				separator = ", "
			}
		}

		sqlStatement := "INSERT INTO courts (" + fields + ") VALUES	(" + values + ") RETURNING id"

		var id2 int
		err := db.QueryRowContext(ctx, sqlStatement).Scan(&id2)
		if err != nil {
			message := "Could not insert into courts"
			f.Errorf(message)
			f.DumpSQLError(err, message, sqlStatement)
			return err
		}
		f.Infof("Inserted court: %d", id2)

		indexes.Courts[id1] = id2
	}

	return nil
}

func insertPlays(ctx context.Context, db *sql.DB, myBackup *backup.Backup, indexes *backup.Indexes) error {
	f := functionInsertPlays

	// Insert plays into the playing table

	for _, play := range myBackup.Playing {
		fields := ""
		values := ""
		separator := ""

		id := indexes.People[play.Person]
		fields = fields + separator + "person"
		values = values + separator + strconv.Itoa(id)
		separator = ", "

		id = indexes.Courts[play.Court]
		fields = fields + separator + "court"
		values = values + separator + strconv.Itoa(id)
		separator = ", "

		sqlStatement := "INSERT INTO playing (" + fields + ") VALUES	(" + values + ")"

		_, err := db.ExecContext(ctx, sqlStatement)
		if err != nil {
			message := "Could not insert into plays"
			f.Errorf(message)
			f.DumpSQLError(err, message, sqlStatement)
			return err
		}
	}

	return nil
}

func insertWaiters(ctx context.Context, db *sql.DB, myBackup *backup.Backup, indexes *backup.Indexes) error {
	f := functionInsertWaiters

	// Insert plays into the waiting table

	for _, waiter := range myBackup.Waiting {

		fields := ""
		separator := ""

		fields = fields + separator + "person"
		separator = ", "
		fields = fields + separator + "start"

		sqlStatement := "INSERT INTO waiting (" + fields + ") VALUES ($1, $2)"

		person := indexes.People[waiter.Person]
		_, err := db.ExecContext(ctx, sqlStatement, person, waiter.Start)
		if err != nil {
			message := "Could not insert into waiting"
			f.Errorf(message)
			f.DumpSQLError(err, message, sqlStatement)
			return err
		}
	}

	return nil
}
