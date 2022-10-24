package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/utils"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

// Position type
type Position struct {
	Index    int      `json:"index"`
	PersonId PersonId `json:"personId"`
}

// Court type
type Court struct {
	ID        int        `json:"id" db:"id"`
	Name      string     `json:"name" db:"name" validate:"required,min=3,max=20"`
	Positions []Position `json:"positions" db:"positions"`
}

// Court type
type PlainCourt struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NullCourt type
type NullCourt struct {
	ID   int
	Name sql.NullString
}

const (
	CourtTable             = "court"
	NumberOfCourtPositions = 4
)

var (
	functionNewCourtFromMap = debug.NewFunction(pkg, "NewCourtFromMap")
	functionUpdateCourt     = debug.NewFunction(pkg, "UpdateCourt")
	functionSaveCourtTx     = debug.NewFunction(pkg, "SaveCourtTx")
	functionSaveCourt       = debug.NewFunction(pkg, "SaveCourt")
	functionListCourtsTx    = debug.NewFunction(pkg, "ListCourtsTx")
	functionListCourts      = debug.NewFunction(pkg, "ListCourts")
	functionLoadCourtTx     = debug.NewFunction(pkg, "LoadCourtTx")
	functionLoadCourt       = debug.NewFunction(pkg, "LoadCourt")
	functionDeleteCourt     = debug.NewFunction(pkg, "DeleteCourt")
	functionDeleteCourtTx   = debug.NewFunction(pkg, "DeleteCourtTx")
)

// NewPerson initialises a Person object
func NewCourt(name string) *Court {
	c := new(Court)
	c.Name = name
	return c
}

func NewCourtFromMap(data *map[string]interface{}) (*Court, error) {
	f := functionNewCourtFromMap
	f.DebugVerbose("")

	name, err := utils.GetStringFromMap("name", data)
	if err != nil {
		return nil, err
	}

	court := NewCourt(name)
	return court, nil
}

func (c *Court) ToPlainCourt() *PlainCourt {

	plainCourt := PlainCourt{
		ID:   c.ID,
		Name: c.Name,
	}

	return &plainCourt
}

// SaveCourt method
func (c *Court) SaveCourt(db *sql.DB) error {
	f := functionSaveCourt
	ctx := context.Background()

	// and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}
	defer EndTransaction(ctx, tx, db, err)

	err = c.SaveCourtTx(ctx, db)
	if err != nil {
		message := "Could not SaveCourt"
		f.DumpError(err, message)
		return err
	}

	return nil
}

// SaveCourtTx writes a new Court to disk and returns the generated id
func (c *Court) SaveCourtTx(ctx context.Context, db *sql.DB) error {
	f := functionSaveCourtTx

	fields := "name"
	values := basic.Quote(c.Name)

	sqlStatement := "INSERT INTO " + CourtTable + " (" + fields + ") VALUES (" + values + ") RETURNING id"
	err := db.QueryRowContext(ctx, sqlStatement).Scan(&c.ID)
	if err != nil {
		message := "Could not insert into " + CourtTable
		d := f.DumpSQLError(err, message, sqlStatement)
		c.Dump(d)
		return err
	}

	return nil
}

// UpdateCourt method
func (c *Court) UpdateCourt(ctx context.Context, db *sql.DB) error {
	f := functionUpdateCourt

	items := "name=" + basic.Quote(c.Name)
	sqlStatement := "UPDATE " + CourtTable + " SET " + items + " WHERE id=" + strconv.Itoa(c.ID)

	_, err := db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not update court"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return err
}

// LoadCourt method
func (c *Court) LoadCourt(db *sql.DB) error {
	f := functionLoadCourt
	ctx := context.Background()

	// and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}
	defer EndTransaction(ctx, tx, db, err)

	err = c.LoadCourtTx(ctx, db)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	return nil
}

// LoadCourtTx returns the Court with the given ID
func (c *Court) LoadCourtTx(ctx context.Context, db *sql.DB) error {
	f := functionLoadCourtTx

	// Query the court
	sqlStatement := "SELECT * FROM " + CourtTable + " WHERE ID=" + strconv.Itoa(c.ID)
	rows, err := db.QueryContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not select all people"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++

		var nc NullCourt
		err := rows.Scan(&nc.ID, &nc.Name)
		if err != nil {
			message := "Could not scan the court"
			f.DumpError(err, message)
		}

		if nc.Name.Valid {
			c.Name = nc.Name.String
		}
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list the courts"
		f.DumpError(err, message)
		return err
	}

	if count == 0 {
		return codeerror.NewNotFound(fmt.Sprintf("Court id %d not found", c.ID))
	} else if count > 1 {
		message := fmt.Sprintf("Found %d courts with id %d", count, c.ID)
		err := codeerror.NewInternalServerError(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

// DeleteCourt removes a court and associated playings
func (c *Court) DeleteCourtTx(db *sql.DB) error {
	f := functionDeleteCourtTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = DeleteCourt(ctx, db, c.ID)
	if err != nil {
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

func DeleteCourt(ctx context.Context, db *sql.DB, courtID int) error {
	f := functionDeleteCourt

	players, err := ListPlayersForCourt(ctx, db, courtID)
	if err != nil {
		message := "Could not delete playings"
		f.DumpError(err, message)
		return err
	}

	for _, player := range players {
		err = MakePlayerWaitTx(ctx, db, player.Person)
		if err != nil {
			message := "Could not make player wait"
			f.DumpError(err, message)
			return err
		}
	}

	// Remove the associated playing
	sqlStatement := "DELETE FROM " + PlayingTable + " WHERE court=" + strconv.Itoa(courtID)
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not delete playings"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Remove the Court
	sqlStatement = "DELETE FROM " + CourtTable + " WHERE ID=" + strconv.Itoa(courtID)
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not delete court"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return nil
}

// ListCourts returns a list of the court IDs
func ListCourts(db *sql.DB) ([]Court, error) {
	f := functionListCourts
	ctx := context.Background()

	// and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}
	defer EndTransaction(ctx, tx, db, err)

	list, err := ListCourtsTx(ctx, db)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// ListCourtsTx returns a list of the court IDs
func ListCourtsTx(ctx context.Context, db *sql.DB) ([]Court, error) {
	f := functionListCourtsTx

	// Query the courts
	returnedFields := []string{`id`, `name`}
	sqlStatement := `SELECT ` + strings.Join(returnedFields, `, `) + ` FROM ` + CourtTable + ` ORDER BY ` + `name`
	rows, err := db.QueryContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not select all from " + CourtTable
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []Court
	for rows.Next() {

		court := Court{}
		court.Positions = make([]Position, 0)

		err := rows.Scan(&court.ID, &court.Name)
		if err != nil {
			message := "Could not scan the court"
			f.DumpError(err, message)
			return nil, err
		}

		players, err := ListPlayersForCourt(ctx, db, court.ID)
		if err != nil {
			message := "Could not list the players on this court"
			f.Errorf(message)
			d := f.DumpError(err, message)

			data, _ := json.MarshalIndent(court, "", "    ")
			d.AddByteArray("court.json", data)

			return nil, err
		}

		for _, player := range players {

			person := FullPerson{ID: player.Person}
			err := person.LoadPersonTx(ctx, db)
			if err != nil {
				message := fmt.Sprintf("Could not load the player [%d]", player.Person)
				d := f.DumpError(err, message)
				d.AddObject("court.json", court)
				d.AddObject("player.json", player)
				return nil, err
			}
			personId := PersonId{ID: player.Person, Knownas: person.Knownas}
			position := Position{Index: player.Position, PersonId: personId}
			court.Positions = append(court.Positions, position)
		}

		list = append(list, court)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all from " + CourtTable
		f.DumpError(err, message)
		return nil, err
	}

	return list, nil
}

// Dump writes the person to a dump file
func (c *Court) Dump(d *debug.Dump) {

	bytearray, err := json.Marshal(c)
	if err != nil {
		return
	}

	title := fmt.Sprintf("court.%d.json", c.ID)
	d.AddByteArray(title, bytearray)
}
