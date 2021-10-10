package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

const (
	// GoodFirstName const
	GoodFirstName = "James"

	// GoodLastName const
	GoodLastName = "Bond"

	// GoodDisplayName const
	GoodDisplayName = "007"

	// GoodUserName const
	GoodUserName = "007"

	// GoodEmail const
	GoodEmail = "007@mi6.gov.uk"

	// GoodPhone const
	GoodPhone = "+44 000 000000"

	// GoodPassword const
	GoodPassword = "TopSecret"

	// AnotherFirstName const
	AnotherFirstName = "Alice"

	// AnotherLastName const
	AnotherLastName = "Smith"

	// AnotherKnownas const
	AnotherKnownas = "Alice"

	// AnotherUserName const
	AnotherUserName = "Ally"

	// AnotherEmail const
	AnotherEmail = "alice@aol.com"

	// AnotherPhone const
	AnotherPhone = "07856 123456"

	// AnotherPassword const
	AnotherPassword = "darkblue"
)

// Logon type
type Logon struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,len=30"`
}

var (
	pkg = debug.NewPackage("model")

	functionMakePlayerPlayTx     = debug.NewFunction(pkg, "MakePlayerPlayTx")
	functionMakePlayerWaitTx     = debug.NewFunction(pkg, "MakePlayerWaitTx")
	functionMakePersonInactiveTx = debug.NewFunction(pkg, "MakePersonInactiveTx")
	functionMakePersonPlayerTx   = debug.NewFunction(pkg, "MakePersonPlayerTx")
	functionMakePersonPlayer     = debug.NewFunction(pkg, "MakePersonPlayer")
	functionPopulate             = debug.NewFunction(pkg, "Populate")
)

// Populate adds a new set of standard records
func Populate(db *sql.DB) error {
	f := functionPopulate
	ctx := context.Background()

	peopleData := []Registration{
		{FirstName: GoodFirstName, LastName: GoodLastName, Knownas: GoodDisplayName, Email: GoodEmail, Phone: GoodPhone, Password: GoodPassword},
		{FirstName: AnotherFirstName, LastName: AnotherLastName, Knownas: AnotherKnownas, Email: AnotherEmail, Phone: AnotherPhone, Password: AnotherPassword},
		{FirstName: "Robert", LastName: "Brown", Knownas: "Bob", Email: "bob@ntl.co.uk", Phone: "012345 123010", Password: "Browneyes"},
		{FirstName: "Charles", LastName: "Winsor", Knownas: "Charlie", Email: "charles@o2.co.uk", Phone: "012345 123011", Password: "hrhcharles"},
		{FirstName: "David", LastName: "Townsend", Knownas: "Dave", Email: "david@bt.co.uk", Phone: "012345 123012", Password: "miltonkeynes"},
		{FirstName: "Edward", LastName: "French", Knownas: "Ed", Email: "immissadda-1167@yopmail.com", Phone: "012345 123013", Password: "romeroandjuliet"},
		{FirstName: "Hana", LastName: "Johnson", Knownas: "Han", Email: "uddobareqi-9086@yopmail.com", Phone: "012345 123014", Password: "tabithathecat"},
		{FirstName: "Annette", LastName: "Mack", Knownas: "Nettie", Email: "benagassuf-0898@yopmail.com", Phone: "012345 123015", Password: "kayleightown"},
		{FirstName: "Karen", LastName: "Curry", Knownas: "Kara", Email: "pyffacisi-2285@yopmail.com", Phone: "012345 123016", Password: "sparkleykeira"},
		{FirstName: "Halima", LastName: "Frazier", Knownas: "Hal", Email: "esunnassuppa-5488@yopmail.com", Phone: "012345 123017", Password: "glitterma"},
		{FirstName: "Laila", LastName: "Mcgrath", Knownas: "La", Email: "enarula-8425@yopmail.com", Phone: "012345 123018", Password: "tinkerham"},
		{FirstName: "Caroline", LastName: "Clarke", Knownas: "Carol", Email: "hossemmibe-4189@yopmail.com", Phone: "012345 123019", Password: "ruificent"},
	}

	peopleIDs := make(map[int]int)
	for i, r := range peopleData {

		p, err := r.ToPerson()
		if err != nil {
			f.Errorf("Could not register person")
			return err
		}

		p.Status = StatusPlayer

		err = p.SavePersonTx(db)
		if err != nil {
			f.Errorf("Could not save person: firstName: %s, lastname: %s, email: %s", p.FirstName, p.LastName, p.Email)
			return err
		}

		err = AddWaiter(ctx, db, p.ID)
		if err != nil {
			f.Errorf("Could not add waiting")
			return err
		}

		peopleIDs[i] = p.ID
	}

	courtData := []struct {
		name string
	}{
		{"A"},
		{"B"},
	}

	courtIDs := make(map[int]int)
	for i, x := range courtData {
		c := Court{Name: x.name}
		err := c.SaveCourt(ctx, db)
		if err != nil {
			message := "Could not save court"
			f.Errorf(message)
			f.DumpError(err, message)
			return err
		}
		courtIDs[i] = c.ID
	}
	return nil
}

// MakePlayerWait moves a person from playing to waiting
func MakePlayerWaitTx(db *sql.DB, personID int) error {
	f := functionMakePlayerWaitTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = MakePlayerWait(ctx, db, personID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func MakePlayerWait(ctx context.Context, db *sql.DB, personID int) error {

	person := FullPerson{ID: personID}
	err := person.LoadPerson(ctx, db)
	if err != nil {
		return codeerror.NewNotFound(fmt.Sprintf("Person [%d] not found", personID))
	}
	if person.Status != StatusPlayer {
		return codeerror.NewBadRequest(fmt.Sprintf("Person [%d] is not a player: state: %s", personID, person.Status))
	}

	err = RemovePlayer(ctx, db, personID)
	if err != nil {
		return err
	}

	err = RemoveWaiter(ctx, db, personID)
	if err != nil {
		return err
	}

	err = AddWaiter(ctx, db, personID)
	if err != nil {
		return err
	}

	return nil
}

// MakePlayerPlaying moves a person from playing to waiting
func MakePlayerPlayTx(db *sql.DB, personID int, courtID int, position int) error {
	f := functionMakePlayerPlayTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = MakePlayerPlay(ctx, db, personID, courtID, position)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func MakePlayerPlay(ctx context.Context, db *sql.DB, personID int, courtID int, position int) error {

	person := FullPerson{ID: personID}
	err := person.LoadPerson(ctx, db)
	if err != nil {
		return codeerror.NewNotFound(fmt.Sprintf("person [%d] not found", personID))
	}
	if person.Status != StatusPlayer {
		return codeerror.NewBadRequest(fmt.Sprintf("person [%d] is not a player: state: %s", personID, person.Status))
	}

	court := Court{ID: courtID}
	err = court.LoadCourt(ctx, db)
	if err != nil {
		return codeerror.NewNotFound(fmt.Sprintf("court [%d] not found", courtID))
	}
	if position < 0 {
		return codeerror.NewBadRequest(fmt.Sprintf("Unexpected position: %d", position))
	}
	if position >= NumberOfCourtPositions {
		return codeerror.NewBadRequest(fmt.Sprintf("Unexpected position: %d", position))
	}

	err = RemovePlayer(ctx, db, personID)
	if err != nil {
		return err
	}

	err = RemoveWaiter(ctx, db, personID)
	if err != nil {
		return err
	}

	err = AddPlayer(ctx, db, personID, courtID, position)
	if err != nil {
		return err
	}

	return nil
}

// MakePersonInactive sets the status of a person to 'inactive'
func MakePersonInactiveTx(db *sql.DB, personID int) error {
	f := functionMakePersonInactiveTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = MakePersonInactive(ctx, db, personID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func MakePersonInactive(ctx context.Context, db *sql.DB, personID int) error {

	person := FullPerson{ID: personID}
	err := person.LoadPerson(ctx, db)
	if err != nil {
		return codeerror.NewNotFound(fmt.Sprintf("person [%d] not found", personID))
	}

	err = RemovePlayer(ctx, db, personID)
	if err != nil {
		return err
	}

	err = RemoveWaiter(ctx, db, personID)
	if err != nil {
		return err
	}

	person.Status = StatusInactive

	err = person.UpdatePerson(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

// MakePersonPlayer sets the status of a person to 'player'
func MakePersonPlayerTx(db *sql.DB, personID int) error {
	f := functionMakePersonPlayerTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	err = MakePersonPlayer(ctx, db, personID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit a new transaction"
		f.Errorf(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

func MakePersonPlayer(ctx context.Context, db *sql.DB, personID int) error {
	f := functionMakePersonPlayer

	players, err := ListPlayersForPerson(ctx, db, personID)
	if err != nil {
		return err
	}

	waiters, err := ListWaitersForPerson(ctx, db, personID)
	if err != nil {
		return err
	}

	if len(players)+len(waiters) > 1 {
		err = codeerror.NewInternalServerError("Unconsistant person")
		d := f.DumpError(err, "")
		data, _ := json.MarshalIndent(struct {
			PersonID        int
			NumberOfPlayers int
			NumberOfWaiters int
		}{
			PersonID:        1,
			NumberOfPlayers: len(players),
			NumberOfWaiters: len(waiters),
		}, "", "    ")
		d.AddByteArray("data.json", data)
		return err
	}

	person := FullPerson{ID: personID}
	err = person.LoadPerson(ctx, db)
	if err != nil {
		return err
	}

	valid := map[string]bool{
		StatusPlayer:   true,
		StatusInactive: true,
	}

	if !valid[person.Status] {
		return codeerror.NewBadRequest(fmt.Sprintf("Cannot change person [%d] from %s to %s state", personID, person.Status, person.Status))
	}

	if person.Status != StatusPlayer {
		person.Status = StatusPlayer
		err = person.UpdatePerson(ctx, db)
		if err != nil {
			return err
		}
	}

	if len(players) > 0 {
		err = RemovePlayer(ctx, db, personID)
		if err != nil {
			return err
		}
	}

	if len(waiters) == 0 {
		err = AddWaiter(ctx, db, personID)
		if err != nil {
			return err
		}
	}

	return nil
}
