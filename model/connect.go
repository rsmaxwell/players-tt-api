package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/rsmaxwell/players-tt-api/internal/config"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionConnect              = debug.NewFunction(pkg, "Connect")
	functionStandardConnect      = debug.NewFunction(pkg, "standardConnect")
	functionBasicConnect         = debug.NewFunction(pkg, "basicConnect")
	functionInitialiseDatabaseTx = debug.NewFunction(pkg, "initialiseDatabaseTx")
	functionInitialiseDatabase   = debug.NewFunction(pkg, "initialiseDatabase")
	functionCreateTables         = debug.NewFunction(pkg, "createTables")
	functionCreateAdminUser      = debug.NewFunction(pkg, "createAdminUser")
	functionDropTables           = debug.NewFunction(pkg, "dropTables")
	functionDropTable            = debug.NewFunction(pkg, "dropTable")
	functionTableExists          = debug.NewFunction(pkg, "tableExists")
	functionDatabaseExists       = debug.NewFunction(pkg, "databaseExists")
	functionCreateDatabase       = debug.NewFunction(pkg, "createDatabase")
	functionDatabaseCheck        = debug.NewFunction(pkg, "databaseCheck")
	functionMakePeople           = debug.NewFunction(pkg, "makePeople")
	functionMakeCourts           = debug.NewFunction(pkg, "makeCourts")
)

const (
	Invalid_Catalog_Name = "3D000"
	Undefined_Table      = "42P01"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	f := functionConnect

	if !strings.EqualFold(os.Getenv("INITIALISE"), "true") {
		db, err := standardConnect(cfg)
		if err != nil {
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
	}

	f.DebugInfo("Connect to postgress (without database)")
	db, err := basicConnect(cfg)
	if err != nil {
		return nil, err
	}

	f.DebugInfo("Check database exists")
	ok, err := databaseExists(db, cfg.Database.DatabaseName)
	if err != nil {
		message := fmt.Sprintf("Problem checking the database '%s'", cfg.Database.DatabaseName)
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}
	if ok {
		f.DebugInfo("Database already exists")
	} else {
		f.DebugInfo("Creating database")
		err = createDatabase(db, cfg.Database.DatabaseName)
		if err != nil {
			message := fmt.Sprintf("Problem checking the database '%s' exists", cfg.Database.DatabaseName)
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}
	}

	db.Close()

	f.DebugInfo("Connect to postgress (with database)")
	db, err = standardConnect(cfg)
	if err != nil {
		return nil, err
	}

	f.DebugInfo("Check database")
	ok, err = databaseCheck(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	if ok {
		f.DebugInfo("Database already initialised")
	} else {
		f.DebugInfo("Re-initialising database")
		err = initialiseDatabaseTx(db)
		if err != nil {
			message := fmt.Sprintf("Problem initialising the database '%s'", cfg.Database.DatabaseName)
			f.Errorf(message)
			f.DumpError(err, message)
			return nil, err
		}
	}

	f.DebugInfo("Re-Checking database")
	ok, err = databaseCheck(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if !ok {
		db.Close()
		return nil, fmt.Errorf("problem initialising database")
	}

	if strings.EqualFold(os.Getenv("POPULATE"), "true") {

		f.DebugInfo("Delete all the records")
		err = DeleteAllRecordsTx(db)
		if err != nil {
			message := "Error delete all the records"
			f.Errorf(message)
			os.Exit(1)
		}

		f.DebugInfo("Populating database with test data - people")
		_, err = MakePeople(db)
		if err != nil {
			f.Errorf("Error making people")
			os.Exit(1)
		}

		f.DebugInfo("Populating database with test data - courts")
		_, err = MakeCourts(db)
		if err != nil {
			f.Errorf("Error making courts")
			os.Exit(1)
		}
	}

	count, err := CheckConistencyTx(db, true)
	if err != nil {
		f.Errorf("Error checking consistency")
		os.Exit(1)
	}

	f.DebugInfo("Made %d database updates", count)

	return db, err
}

func standardConnect(cfg *config.Config) (*sql.DB, error) {
	f := functionStandardConnect

	// Connect to postgres (with database)
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

func basicConnect(cfg *config.Config) (*sql.DB, error) {
	f := functionBasicConnect

	// Connect to postgres (no database)
	f.DebugInfo("Failed database check")
	f.DebugInfo("Connect to postgres (no database)")
	driverName := cfg.DriverName()
	connectionString := cfg.ConnectionStringBasic()
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

	err = createAdminUser(ctx, db)
	if err != nil {
		return nil
	}

	return nil
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

	peopleData := []Registration{
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

		p.Status = StatusAdmin

		err = p.SavePerson(ctx, db)
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

	f.DebugInfo("Successfully created Tables\n")
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

// Person type
type PersonData struct {
	Data   Registration
	Status string
}

func MakePeople(db *sql.DB) (map[int]int, error) {
	f := functionMakePeople

	peopleData := []PersonData{
		{Data: Registration{FirstName: "James", LastName: "Bond", Knownas: "007", Email: "007@mi6.gov.uk", Phone: "01632 960573", Password: "TopSecret123"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Alice", LastName: "Frombe", Knownas: "ali", Email: "ali@mikymouse.com", Phone: "01632 960372", Password: "ali1234567"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Tom", LastName: "Smith", Knownas: "tom", Email: "tom@hotmail.com", Phone: "01632 960512", Password: "tom12378909876"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Sandra", LastName: "Smythe", Knownas: "tom", Email: "sandra@hotmail.com", Phone: "01632 960966", Password: "sandra12334567"}, Status: StatusInactive},
		{Data: Registration{FirstName: "George", LastName: "Washington", Knownas: "george", Email: "george@hotmail.com", Phone: "01632 960278", Password: "george789"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Margret", LastName: "Tiffington", Knownas: "maggie", Email: "marg@hotmail.com", Phone: "01632 960165", Password: "magie876"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "James", LastName: "Ernest", Knownas: "jamie", Email: "jamie@ntlworld.com", Phone: "01632 960757", Password: "jamie5293645284"}, Status: StatusInactive},
		{Data: Registration{FirstName: "Elizabeth", LastName: "Tudor", Knownas: "liz", Email: "liz@buck.palice.com", Phone: "01632 960252", Password: "liz1756453423"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Dick", LastName: "Whittington", Knownas: "dick", Email: "dick@ntlworld.com", Phone: "01746 352413", Password: "dick3296846734524"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Victoria", LastName: "Hempworth", Knownas: "vickie", Email: "vickie@waitrose.com", Phone: "0195 76863241", Password: "vickie846"}, Status: StatusPlayer},

		{Data: Registration{FirstName: "Shanika", LastName: "Pierre", Knownas: "pete", Email: "IcyGamer@gmail.com", Phone: "01632 960576", Password: "Top12345Secret"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Wanangwa", LastName: "Czajkowski", Knownas: "wan", Email: "torphy.dayana@dicki.com", Phone: "01632 960628", Password: "ali12387654"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Cormac", LastName: "Dwight", Knownas: "cor", Email: "adela.kunze@schmitt.com", Phone: "01632 960026", Password: "tom123frgthyj"}, Status: StatusSuspended},
		{Data: Registration{FirstName: "Ramóna", LastName: "Jonker", Knownas: "ram", Email: "ariel07@hotmail.com", Phone: "01632 960801", Password: "sandra123frr"}, Status: StatusSuspended},
		{Data: Registration{FirstName: "Quinctilius", LastName: "Jack", Knownas: "qui", Email: "kara.johnston@runte.com", Phone: "01632 960334", Password: "george789ed5"}, Status: StatusInactive},
		{Data: Registration{FirstName: "Radu", LastName: "Godfrey", Knownas: "rad", Email: "ella.vonrueden@kuhic.com", Phone: "01632 960450", Password: "magie87689ilom"}, Status: StatusSuspended},
		{Data: Registration{FirstName: "Aleksandrina", LastName: "Couture", Knownas: "ale", Email: "archibald.stark@hotmail.com", Phone: "01632 960928", Password: "jamie529re5gb"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Catrin", LastName: "Wooldridge", Knownas: "cat", Email: "sauer.luciano@hotmail.com", Phone: "01632 960126", Password: "liz14rdgujmbvr43"}, Status: StatusSuspended},
		{Data: Registration{FirstName: "Souleymane", LastName: "Walter", Knownas: "sou", Email: "damon.toy@swaniawski.com", Phone: "01632 960403", Password: "dick3287uyh5fredw"}, Status: StatusPlayer},
		{Data: Registration{FirstName: "Dorotėja", LastName: "Antúnez", Knownas: "dor", Email: "omante@marks.com", Phone: "01632 961252", Password: "vickie846y6"}, Status: StatusPlayer},
	}

	peopleIDs := make(map[int]int)
	for i, r := range peopleData {

		p, err := r.Data.ToPerson()
		if err != nil {
			message := "Could not register person"
			f.Errorf(message)
			f.DumpError(err, message)
			os.Exit(1)
		}

		p.Status = r.Status

		err = p.SavePersonTx(db)
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
		f.DebugInfo("    Password:  %s", r.Data.Password)
		f.DebugInfo("    Hash:      %s", p.Hash)
		f.DebugInfo("    Status:    %s", p.Status)
	}

	return peopleIDs, nil
}

// Court type
type CourtData struct {
	Name string
}

func MakeCourts(db *sql.DB) (map[int]int, error) {
	f := functionMakeCourts

	courtsData := []CourtData{
		{Name: "A"},
		{Name: "B"},
	}

	courtIDs := make(map[int]int)
	for i, c := range courtsData {

		court := Court{Name: c.Name}

		err := court.SaveCourtTx(db)
		if err != nil {
			message := fmt.Sprintf("Could not save court: Name: %s", court.Name)
			f.Errorf(message)
			f.DumpError(err, message)
			os.Exit(1)
		}

		courtIDs[i] = court.ID

		f.DebugInfo("Added court:")
		f.DebugInfo("    ID:    %d", court.ID)
		f.DebugInfo("    Name:  %s", court.Name)
	}

	return courtIDs, nil
}
