package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionCheckConistencyTx     = debug.NewFunction(pkg, "CheckConistencyTx")
	functionCheckConistency       = debug.NewFunction(pkg, "CheckConistency")
	functionCheckConistencyPerson = debug.NewFunction(pkg, "CheckConistencyPerson")
)

func CheckConistencyTx(db *sql.DB, fix bool) (int, error) {
	f := functionCheckConistencyTx
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		message := "Could not begin a new transaction"
		f.DumpError(err, message)
		return 0, err
	}

	count, err := CheckConistency(ctx, db, fix)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		message := "Could not commit the transaction"
		f.DumpError(err, message)
	}

	return count, nil
}

func CheckConistency(ctx context.Context, db *sql.DB, fix bool) (int, error) {
	f := functionCheckConistency

	list, err := ListPeople(ctx, db, "")
	if err != nil {
		message := "Could not list people"
		f.Errorf(message)
		f.DumpError(err, message)
		return 0, err
	}

	total := 0

	for _, person := range list {

		count, err := person.CheckConistencyPerson(ctx, db, fix)
		if err != nil {
			message := fmt.Sprintf("Problem checking consistany for person: [%d: %s]", person.ID, person.Knownas)
			f.Errorf(message)
			f.DumpError(err, message)
			return 0, err
		}

		total = total + count
	}

	return total, nil
}

func (person *FullPerson) CheckConistencyPerson(ctx context.Context, db *sql.DB, fix bool) (int, error) {
	f := functionCheckConistencyPerson

	count := 0

	waiters, err := ListWaitersForPerson(ctx, db, person.ID)
	if err != nil {
		message := fmt.Sprintf("Could not list waiters for person: [%d: %s]", person.ID, person.Knownas)
		f.Errorf(message)
		f.DumpError(err, message)
		return 0, err
	}

	players, err := ListPlayersForPerson(ctx, db, person.ID)
	if err != nil {
		message := fmt.Sprintf("Could not list players for person: [%d: %s]", person.ID, person.Knownas)
		f.Errorf(message)
		f.DumpError(err, message)
		return 0, err
	}

	if person.Status == StatusPlayer {
		if len(waiters) < 1 {
			if len(players) < 1 {

				count++

				if fix {
					f.DebugError(fmt.Sprintf("Adding waiter record for person [%d: %s]", person.ID, person.Knownas))
					err := AddWaiter(ctx, db, person.ID)
					if err != nil {
						message := fmt.Sprintf("Could not add waiter: [%d: %s]", person.ID, person.Knownas)
						f.Errorf(message)
						f.DumpError(err, message)
						return 0, err
					}
				} else {
					f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is a player but has no waiter or player records", person.ID, person.Knownas))
				}
			} else if len(players) > 1 {

				count++

				if fix {
					f.DebugError(fmt.Sprintf("Removing player record and adding waiter record for person [%d: %s]", person.ID, person.Knownas))

					err = RemovePlayer(ctx, db, person.ID)
					if err != nil {
						message := fmt.Sprintf("Could not remove player: id: [%d: %s]", person.ID, person.Knownas)
						f.Errorf(message)
						f.DumpError(err, message)
						return 0, err
					}

					err := AddWaiter(ctx, db, person.ID)
					if err != nil {
						message := fmt.Sprintf("Could not add waiter: id: [%d: %s]", person.ID, person.Knownas)
						f.Errorf(message)
						f.DumpError(err, message)
						return 0, err
					}
				} else {
					f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is a player and has %d player records", person.ID, person.Knownas, len(players)))
				}
			}
		} else if len(waiters) > 1 {

			count++

			if fix {
				f.DebugError(fmt.Sprintf("Removing waiter and player records, then adding new waiter record for person [%d: %s]", person.ID, person.Knownas))

				err = RemoveWaiter(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not remove waiter: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}

				err = RemovePlayer(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not remove player: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}

				err := AddWaiter(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not add waiter: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}
			} else {
				f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is a player but has %d waiter records", person.ID, person.Knownas, len(waiters)))
			}
		} else if len(players) < 1 {
			// NOP
		} else {

			count++

			if fix {
				f.DebugError(fmt.Sprintf("Removing player record for person [%d: %s]", person.ID, person.Knownas))

				err = RemovePlayer(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not remove player: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}
			} else {
				f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is a player but has 1 waiter record and %d player records", person.ID, person.Knownas, len(players)))
			}
		}
	} else {

		if len(waiters) > 0 {

			count++

			if fix {
				f.DebugError(fmt.Sprintf("Removing waiter record for person [%d: %s]", person.ID, person.Knownas))

				err = RemoveWaiter(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not remove waiter: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}
			} else {
				f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is not a player but has %d waiter records", person.ID, person.Knownas, len(waiters)))
			}
		}

		if len(players) > 0 {

			count++

			if fix {
				f.DebugError(fmt.Sprintf("Removing player record for person [%d: %s]", person.ID, person.Knownas))

				err = RemovePlayer(ctx, db, person.ID)
				if err != nil {
					message := fmt.Sprintf("Could not remove player: [%d: %s]", person.ID, person.Knownas)
					f.Errorf(message)
					f.DumpError(err, message)
					return 0, err
				}
			} else {
				f.DebugError(fmt.Sprintf("Inconsistant data: person [%d: %s] is not a player but has %d player records", person.ID, person.Knownas, len(players)))
			}
		}
	}

	return count, nil
}
