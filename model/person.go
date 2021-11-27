package model

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/utils"
)

// LimitedPerson type
type Person struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Knownas   string `json:"knownas"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
}

// Person type
type FullPerson struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstname" validate:"required,min=3,max=20"`
	LastName  string `json:"lastname" validate:"required,min=3,max=20"`
	Knownas   string `json:"knownas" validate:"required,min=3,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,min=3,max=20"`
	Hash      []byte `json:"hash"`
	Status    string `json:"status"`
}

// NullPerson type
type NullPerson struct {
	ID        int            `db:"id"`
	FirstName sql.NullString `db:"firstname"`
	LastName  sql.NullString `db:"lastname"`
	Knownas   sql.NullString `db:"knownas"`
	Email     sql.NullString `db:"email"`
	Phone     sql.NullString `db:"phone"`
	Hash      sql.NullString `db:"hash"`
	Status    sql.NullString `db:"status"`
}

const (
	// PersonTable is the name of the person table
	PersonTable = "person"
)

var (
	functionNewPersonFromMap  = debug.NewFunction(pkg, "NewPersonFromMap")
	functionUpdatePerson      = debug.NewFunction(pkg, "UpdatePerson")
	functionSavePersonTx      = debug.NewFunction(pkg, "SavePersonTx")
	functionSavePerson        = debug.NewFunction(pkg, "SavePerson")
	functionFindPersonByEmail = debug.NewFunction(pkg, "FindPersonByEmail")
	functionListPeopleTx      = debug.NewFunction(pkg, "ListPeopleTx")
	functionListPeople        = debug.NewFunction(pkg, "ListPeople")
	functionLoadPersonTx      = debug.NewFunction(pkg, "LoadPersonTx")
	functionLoadPerson        = debug.NewFunction(pkg, "LoadPerson")
	functionDeletePersonTx    = debug.NewFunction(pkg, "DeletePersonTx")
	functionDeletePerson      = debug.NewFunction(pkg, "DeletePerson")
	functionAuthenticate      = debug.NewFunction(pkg, "Authenticate")
	functionCheckPassword     = debug.NewFunction(pkg, "CheckPassword")
)

const (
	// StatusAdmin constant
	StatusAdmin = "admin"

	// StatusPlayer constant
	StatusPlayer = "player"

	// StatusInactive constant
	StatusInactive = "inactive"

	// StatusSuspended constant
	StatusSuspended = "suspended"
)

var (
	// AllStates lists all the states
	AllStates []string
)

func init() {
	// AllRoles lists all the roles
	AllStates = []string{StatusAdmin, StatusPlayer, StatusInactive, StatusSuspended}
}

// NewPerson initialises a Person object
func NewPerson(firstname string, lastname string, knownas string, email string, phone string, hash []byte) *FullPerson {
	p := new(FullPerson)
	p.FirstName = firstname
	p.LastName = lastname
	p.Knownas = knownas
	p.Email = email
	p.Phone = phone
	p.Hash = hash
	p.Status = StatusSuspended
	return p
}

func NewPersonFromMap(data *map[string]interface{}) (*Registration, error) {
	f := functionNewPersonFromMap
	f.DebugVerbose("")

	firstname, err := utils.GetStringFromMap("firstname", data)
	if err != nil {
		return nil, err
	}

	lastname, err := utils.GetStringFromMap("lastname", data)
	if err != nil {
		return nil, err
	}

	email, err := utils.GetStringFromMap("email", data)
	if err != nil {
		return nil, err
	}

	knownas, err := utils.GetStringFromMap("knownas", data)
	if err != nil {
		return nil, err
	}

	phone, err := utils.GetStringFromMap("phone", data)
	if err != nil {
		return nil, err
	}

	password, err := utils.GetStringFromMap("password", data)
	if err != nil {
		return nil, err
	}

	person := NewRegistration(firstname, lastname, knownas, email, phone, password)
	return person, nil
}

// DeletePerson removes a person and associated waiters and playings
func (p *FullPerson) SavePersonTx(db *sql.DB) error {
	f := functionSavePersonTx
	ctx := context.Background()

	// Create a new context, and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = p.SavePerson(ctx, db)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
	}

	return nil
}

// SavePerson writes a new Person to disk and returns the generated id
func (p *FullPerson) SavePerson(ctx context.Context, db *sql.DB) error {
	f := functionSavePerson

	fields := "firstname, lastname, knownas, email, phone, hash, status"
	values := "$1, $2, $3, $4, $5, $6, $7"
	sqlStatement := "INSERT INTO " + PersonTable + " (" + fields + ") VALUES (" + values + ") RETURNING id"

	err := db.QueryRowContext(ctx, sqlStatement, p.FirstName, p.LastName, p.Knownas, p.Email, p.Phone, hex.EncodeToString(p.Hash), p.Status).Scan(&p.ID)
	if err != nil {
		pgerr, ok := err.(*pgconn.PgError)
		if ok {
			if pgerr.Code == "23505" {
				return err
			}
		}

		message := "Could not insert into " + PersonTable
		d := f.DumpSQLError(err, message, sqlStatement)
		p.Dump(d)
		return err
	}

	return nil
}

func (p *FullPerson) UpdatePerson(ctx context.Context, db *sql.DB) error {
	f := functionUpdatePerson

	fields := "firstname=$1, lastname=$2, knownas=$3, email=$4, phone=$5, hash=$6, status=$7"
	sqlStatement := "UPDATE " + PersonTable + " SET " + fields + " WHERE id=" + strconv.Itoa(p.ID)
	_, err := db.ExecContext(ctx, sqlStatement, p.FirstName, p.LastName, p.Knownas, p.Email, p.Phone, hex.EncodeToString(p.Hash), p.Status)
	if err != nil {
		message := "Could not update person"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return err
}

// LoadPerson returns the Person with the given ID
func (p *FullPerson) LoadPersonTx(db *sql.DB) error {
	f := functionLoadPersonTx
	ctx := context.Background()

	// Create a new context, and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = p.LoadPerson(ctx, db)
	if err != nil {
		tx.Rollback()
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

func (p *FullPerson) LoadPerson(ctx context.Context, db *sql.DB) error {
	f := functionLoadPerson

	// Query the person
	fields := "firstname, lastname, knownas, email, phone, hash, status"
	sqlStatement := "SELECT " + fields + " FROM " + PersonTable + " WHERE id=$1"
	rows, err := db.QueryContext(ctx, sqlStatement, p.ID)
	if err != nil {
		message := "Could not select all people"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++

		var np NullPerson
		err := rows.Scan(&np.FirstName, &np.LastName, &np.Knownas, &np.Email, &np.Phone, &np.Hash, &np.Status)
		if err != nil {
			message := "Could not scan the person"
			f.DumpError(err, message)
			return err
		}

		if np.FirstName.Valid {
			p.FirstName = np.FirstName.String
		}

		if np.LastName.Valid {
			p.LastName = np.LastName.String
		}

		if np.Knownas.Valid {
			p.Knownas = np.Knownas.String
		}

		if np.Email.Valid {
			p.Email = np.Email.String
		}

		if np.Phone.Valid {
			p.Phone = np.Phone.String
		}

		if np.Hash.Valid {
			p.Hash, err = hex.DecodeString(np.Hash.String)
			if err != nil {
				message := "Could not scan the Hash HexString"
				f.DumpError(err, message)
				return err
			}
		}

		if np.Status.Valid {
			p.Status = np.Status.String
		}
	}
	err = rows.Err()
	if err != nil {
		message := "Could not query the person"
		f.DumpError(err, message)
		return err
	}

	if count == 0 {
		f.Infof("sqlStatement: %s", sqlStatement)
		return codeerror.NewNotFound(fmt.Sprintf("Person ID %d not found", p.ID))
	} else if count > 1 {
		message := fmt.Sprintf("Found %d people with id %d", count, p.ID)
		err := codeerror.NewInternalServerError(message)
		f.DumpError(err, message)
		return err
	}

	return nil
}

// DeletePerson removes a person and associated waiters and playings
func (p *FullPerson) DeletePersonTx(db *sql.DB) error {
	f := functionDeletePersonTx
	ctx := context.Background()

	// Create a new context, and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = DeletePerson(ctx, db, p.ID)
	if err != nil {
		tx.Rollback()
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

func DeletePerson(ctx context.Context, db *sql.DB, personID int) error {
	f := functionDeletePerson

	// Remove the associated waiters
	sqlStatement := "DELETE FROM " + WaitingTable + " WHERE person=" + strconv.Itoa(personID)
	_, err := db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not delete waiters"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Remove the associated playing
	sqlStatement = "DELETE FROM " + PlayingTable + " WHERE person=" + strconv.Itoa(personID)
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not delete playings"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	// Remove the Person
	sqlStatement = "DELETE FROM " + PersonTable + " WHERE ID=" + strconv.Itoa(personID) + " AND status != '" + StatusAdmin + "'"
	_, err = db.ExecContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not delete person"
		f.DumpSQLError(err, message, sqlStatement)
		return err
	}

	return nil
}

// FindPersonByEmail function
func FindPersonByEmail(ctx context.Context, db *sql.DB, email string) (*FullPerson, error) {
	f := functionFindPersonByEmail

	// Query the people
	fields := "id, firstname, lastname, knownas, email, phone, hash, status"
	where := `email=$1`
	sqlStatement := `SELECT ` + fields + ` FROM ` + PersonTable + ` WHERE ` + where

	f.DebugVerbose("sqlStatement: %s", sqlStatement)
	rows, err := db.QueryContext(ctx, sqlStatement, email)
	if err != nil {
		message := "Could not select all from " + PersonTable
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var arrayOfPeople []FullPerson
	for rows.Next() {

		var p FullPerson
		var hexstring string
		err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Knownas, &p.Email, &p.Phone, &hexstring, &p.Status)
		if err != nil {
			message := "Could not scan the person"
			f.DumpError(err, message)
			return nil, err
		}

		p.Hash, err = hex.DecodeString(hexstring)
		if err != nil {
			message := "Could not decode hextring: " + hexstring
			f.DumpError(err, message)
			return nil, err
		}

		// fmt.Printf("    FirstName: %s\n", p.FirstName)
		// fmt.Printf("    LastName:  %s\n", p.LastName)
		// fmt.Printf("    Knownas:  %s\n", p.Knownas)
		// fmt.Printf("    email:     %s\n", p.Email)
		// fmt.Printf("    hexstring: %s\n", hexstring)
		// fmt.Printf("    hash:      %v\n", p.Hash)

		arrayOfPeople = append(arrayOfPeople, p)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all from " + PersonTable
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}

	if len(arrayOfPeople) <= 0 {
		err := codeerror.NewNotFound(fmt.Sprintf("Person not found: email:%s", email))
		return nil, err
	}

	if len(arrayOfPeople) > 1 {
		message := fmt.Sprintf("Too many matches. email:%s, count:%d", email, len(arrayOfPeople))
		err := codeerror.NewNotFound(message)
		f.DumpError(err, message)
		return nil, err
	}

	return &arrayOfPeople[0], nil
}

// ListPeople function
func ListPeopleTx(db *sql.DB, whereClause string) ([]FullPerson, error) {
	f := functionListPeopleTx
	ctx := context.Background()

	// Create a new context, and begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return nil, err
	}

	listOfPeople, err := ListPeople(ctx, db, whereClause)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
		return nil, err
	}

	return listOfPeople, nil
}

// ListPeople returns a list of the people IDs
func ListPeople(ctx context.Context, db *sql.DB, whereClause string) ([]FullPerson, error) {
	f := functionListPeople

	// Query the people
	fields := "id, firstname, lastname, knownas, email, phone, hash, status"
	sqlStatement := `SELECT ` + fields + ` FROM ` + PersonTable + ` ` + whereClause + ` ORDER BY ` + `knownas`
	rows, err := db.QueryContext(ctx, sqlStatement)
	if err != nil {
		message := "Could not select all from " + PersonTable
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}
	defer rows.Close()

	var list []FullPerson
	for rows.Next() {

		var p FullPerson
		var hexstring string
		err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Knownas, &p.Email, &p.Phone, &hexstring, &p.Status)
		if err != nil {
			message := "Could not scan the person"
			f.DumpError(err, message)
			return nil, err
		}

		p.Hash, err = hex.DecodeString(hexstring)
		if err != nil {
			message := "Could not decode hextring: " + hexstring
			f.DumpError(err, message)
			return nil, err
		}

		// fmt.Printf("    FirstName: %s\n", p.FirstName)
		// fmt.Printf("    LastName:  %s\n", p.LastName)
		// fmt.Printf("    Knownas:  %s\n", p.Knownas)
		// fmt.Printf("    email:     %s\n", p.Email)
		// fmt.Printf("    hexstring: %s\n", hexstring)
		// fmt.Printf("    hash:      %v\n", p.Hash)

		list = append(list, p)
	}
	err = rows.Err()
	if err != nil {
		message := "Could not list all from " + PersonTable
		f.DumpSQLError(err, message, sqlStatement)
		return nil, err
	}

	return list, nil
}

// Authenticate method
func (p *FullPerson) Authenticate(db *sql.DB, password string) error {
	f := functionAuthenticate

	err := p.checkPassword(password)
	if err != nil {
		f.DebugVerbose("password check failed for person [%d]", p.ID)
		return codeerror.NewUnauthorized("Not Authorized")
	}

	err = p.CanLogin()
	if err != nil {
		f.DebugVerbose("person [%d] not authorized to login", p.ID)
		return codeerror.NewForbidden("Forbidden")
	}

	return nil
}

// CheckPassword checks the validity of the password
func (p *FullPerson) checkPassword(password string) error {
	f := functionCheckPassword

	// fmt.Printf("    FirstName: %s\n", p.FirstName)
	// fmt.Printf("    LastName:  %s\n", p.LastName)
	// fmt.Printf("    email:     %s\n", p.Email)
	// fmt.Printf("    hash:      %v\n", p.Hash)
	// fmt.Printf("    hash:      %s\n", hex.EncodeToString(p.Hash))

	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(password))
	if err != nil {
		message := fmt.Sprintf("The password %s was invalid for the user with email: %s", password, p.Email)
		d := f.DumpError(err, message)
		d.AddString("hash.txt", hex.EncodeToString(p.Hash))
		return err
	}
	return nil
}

// CanLogin checks the user is allowed to login
func (p *FullPerson) CanLogin() error {

	if p.Status == StatusAdmin {
		return nil
	}
	if p.Status == StatusPlayer {
		return nil
	}
	if p.Status == StatusInactive {
		return nil
	}

	return fmt.Errorf("not Authorized")
}

// CanEditCourt checks the user is allowed update a court
func (p *FullPerson) CanEditCourt() error {

	if p.Status == StatusAdmin {
		return nil
	}
	if p.Status == StatusPlayer {
		return nil
	}

	return fmt.Errorf("not Authorized")
}

// CanGetMetrics checks the user is allowed get the metrics
func (p *FullPerson) CanGetMetrics() error {

	if p.Status == StatusAdmin {
		return nil
	}
	if p.Status == StatusPlayer {
		return nil
	}
	if p.Status == StatusInactive {
		return nil
	}

	return fmt.Errorf("not Authorized")
}

// CanEditOtherPeople checks the user is allowed update a court
func (p *FullPerson) CanEditOtherPeople() error {

	if p.Status == StatusAdmin {
		return nil
	}
	if p.Status == StatusPlayer {
		return nil
	}
	if p.Status == StatusInactive {
		return nil
	}

	return fmt.Errorf("not Authorized")
}

// CanEditSelf checks the user is allowed update a court
func (p *FullPerson) CanEditSelf() error {

	if p.Status == StatusAdmin {
		return nil
	}
	if p.Status == StatusPlayer {
		return nil
	}
	if p.Status == StatusInactive {
		return nil
	}

	return fmt.Errorf("not Authorized")
}

// ToLimited converts a person to a Limited person
func (p *FullPerson) ToLimited() *Person {
	lp := &Person{
		ID:        p.ID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Knownas:   p.Knownas,
		Email:     p.Email,
		Phone:     p.Phone,
		Status:    p.Status,
	}
	return lp
}

// ToLimitedPerson converts a NullPerson to a Limited person
// func (np *NullPerson) xToLimitedPerson() *Person {

// 	lp := Person{}

// 	if np.FirstName.Valid {
// 		lp.FirstName = np.FirstName.String
// 	}

// 	if np.LastName.Valid {
// 		lp.LastName = np.LastName.String
// 	}

// 	if np.Knownas.Valid {
// 		lp.Knownas = np.Knownas.String
// 	}

// 	if np.Email.Valid {
// 		lp.Email = np.Email.String
// 	}

// 	if np.Phone.Valid {
// 		lp.Phone = np.Phone.String
// 	}

// 	if np.Status.Valid {
// 		lp.Status = np.Status.String
// 	}

// 	return &lp
// }

// Dump writes the person to a dump file
func (p *FullPerson) Dump(d *debug.Dump) {

	bytearray, err := json.Marshal(p)
	if err != nil {
		return
	}

	title := fmt.Sprintf("person.%d.json", p.ID)
	d.AddByteArray(title, bytearray)
}
