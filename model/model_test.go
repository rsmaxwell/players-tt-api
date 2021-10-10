package model

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
)

func TestPeople(t *testing.T) {

	teardown, db, _ := Setup(t)
	defer teardown(t)

	ctx := context.Background()

	err := DeleteAllRecords(ctx, db)
	if err != nil {
		t.Log("Could not setup the model")
		t.FailNow()
	}

	err = Populate(db)
	if err != nil {
		t.Log("Could not populate")
		t.FailNow()
	}

	listOfCourts, err := ListCourtsTx(db)
	if err != nil {
		t.Log("Could not list the courts")
		t.FailNow()
	}
	if len(listOfCourts) == 0 {
		t.Log("Could not find any courts")
		t.FailNow()
	}
	var c Court
	// c.ID = listOfCourts[0]
	// err = c.LoadCourt(db)
	// if err != nil {
	// 	t.Log("Could not load court")
	// 	t.FailNow()
	// }

	listOfWaiters, err := ListWaiters(ctx, db)
	if err != nil {
		t.Log("Could not get the first waiter")
		t.FailNow()
	}
	if len(listOfWaiters) == 0 {
		t.Log("Could not find any waiters")
		t.FailNow()
	}
	var p FullPerson
	p.ID = listOfWaiters[0].Person
	err = p.LoadPerson(ctx, db)
	if err != nil {
		t.Log("Could not get the first waiter")
		t.FailNow()
	}

	err = RemoveWaiter(ctx, db, p.ID)
	if err != nil {
		t.Log("Could not remove waiter")
		t.FailNow()
	}

	err = AddWaiter(ctx, db, p.ID)
	if err != nil {
		t.Log("Could not make a person into a player")
		t.FailNow()
	}

	err = MakePersonInactive(ctx, db, p.ID)
	if err != nil {
		t.Log("Could not make a person inactive")
		t.FailNow()
	}

	p.FirstName = "smersh"
	p.LastName = "Bomb"

	err = p.UpdatePerson(ctx, db)
	if err != nil {
		t.Log("Could not update person")
		t.FailNow()
	}

	var p2 FullPerson
	p2.ID = p.ID
	err = p2.LoadPerson(ctx, db)
	if err != nil {
		t.Log("Could not load person")
		t.FailNow()
	}

	p2.FirstName = "xxxxx"
	p2.Email = "fabdelkader.browx@balaways.com"
	p2.Phone = "+44 012 098765"
	err = p2.SavePersonTx(db)
	if err != nil {
		message := "Could not save person"
		t.Log(message)
		t.Log(err)
		t.FailNow()
	}

	c.Name = "AAAAA"

	err = c.UpdateCourt(ctx, db)
	if err != nil {
		t.Log("Could not update court")
		t.FailNow()
	}

	err = c.DeleteCourtTx(db)
	if err != nil {
		t.Log("Could not delete court")
		t.FailNow()
	}

	err = p.DeletePersonTx(db)
	if err != nil {
		t.Log("Could not delete person")
		t.FailNow()
	}
	err = p2.DeletePersonTx(db)
	if err != nil {
		t.Log("Could not delete person")
		t.FailNow()
	}
}
