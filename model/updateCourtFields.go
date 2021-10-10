package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionUpdateCourtFields   = debug.NewFunction(pkg, "UpdateCourtFields")
	functionUpdateCourtFieldsTx = debug.NewFunction(pkg, "UpdateCourtFieldsTx")
)

// UpdateCourt method
func UpdateCourtFieldsTx(db *sql.DB, courtID int, fields map[string]interface{}) error {
	f := functionUpdateCourtFieldsTx
	ctx := context.Background()

	// Begin a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return err
	}

	err = UpdateCourtFields(ctx, db, courtID, fields)
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

func UpdateCourtFields(ctx context.Context, db *sql.DB, courtID int, fields map[string]interface{}) error {
	f := functionUpdateCourtFields

	c := Court{ID: courtID}
	err := c.LoadCourt(ctx, db)
	if err != nil {
		message := fmt.Sprintf("could not load court: %d", courtID)
		f.DebugVerbose(message)
		d := f.DumpError(err, message)
		d.AddObject("fields", fields)
		return codeerror.NewInternalServerError(message)
	}

	if val, ok := fields["name"]; ok {
		c.Name, ok = val.(string)
		if !ok {
			message := fmt.Sprintf("unexpected type for [%s]: %v", "name", val)
			f.DebugVerbose(message)
			f.DumpError(err, message)
			return codeerror.NewInternalServerError(message)
		}
	}

	err = c.UpdateCourt(ctx, db)
	if err != nil {
		message := fmt.Sprintf("problem updating court: %d", courtID)
		f.DebugVerbose(message)
		f.DumpError(err, message)
		return codeerror.NewInternalServerError(message)
	}

	return nil
}
