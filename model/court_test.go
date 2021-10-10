package model

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
)

func TestCourts(t *testing.T) {
	teardown, db, _ := Setup(t)
	defer teardown(t)

	name1 := "A"
	name2 := "Number 1"
	name3 := "xxxxx"

	ctx := context.Background()

	c := Court{Name: name1}
	err := c.SaveCourt(ctx, db)
	if err != nil {
		t.Log("Could not create new court")
		t.FailNow()
	}
	c.Check(ctx, t, db, name1)

	c.Name = name2
	err = c.UpdateCourt(ctx, db)
	if err != nil {
		t.Log("Could not update court")
		t.FailNow()
	}
	c.Check(ctx, t, db, name2)

	var c2 = Court{ID: c.ID}
	err = c2.LoadCourt(ctx, db)
	if err != nil {
		t.Log("Could not load court")
		t.FailNow()
	}
	c2.Check(ctx, t, db, name2)

	c2.Name = name3
	err = c2.SaveCourt(ctx, db)
	if err != nil {
		t.Log("Could not save court")
		t.FailNow()
	}
	c2.Check(ctx, t, db, name3)

	err = c.DeleteCourtTx(db)
	if err != nil {
		t.Log("Could not delete court")
		t.FailNow()
	}
	err = c2.DeleteCourtTx(db)
	if err != nil {
		t.Log("Could not delete court")
		t.FailNow()
	}
}

func (c *Court) Check(ctx context.Context, t *testing.T, db *sql.DB, name string) error {
	err := c.LoadCourt(ctx, db)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if c.Name != name {
		t.Logf("Unexpected name. expected: '%s' actual: '%s'", name, c.Name)
		t.FailNow()
	}

	return nil
}
