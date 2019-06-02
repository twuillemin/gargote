package db

import (
	"errors"
	"github.com/hashicorp/go-memdb"
)

// db is the pointer to the in memory database
var db *memdb.MemDB

// ActionEntry is the format of the data for recording the execution of a single action
type ActionEntry struct {
	TestIndex    int
	StageIndex   int
	TryNumber    int
	ActionIndex  int
	TimeNano     int
	DurationNano int
	Success      bool
}

// CreateDatabase initializes the in memory database
//
// Return an error if the database can not be initialized
func CreateDatabase() error {

	compoundIndex := memdb.CompoundIndex{
		Indexes: []memdb.Indexer{
			&memdb.IntFieldIndex{Field: "TestIndex"},
			&memdb.IntFieldIndex{Field: "StageIndex"},
			&memdb.IntFieldIndex{Field: "TryNumber"},
			&memdb.IntFieldIndex{Field: "ActionIndex"},
		},
		AllowMissing: false,
	}

	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"action": {
				Name: "action",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &compoundIndex,
					},
					"time_index": {
						Name:    "time_index",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "TimeNano"},
					},
				},
			},
		},
	}

	// Create a new data base
	newDatabase, err := memdb.NewMemDB(schema)
	if err != nil {
		return err
	}

	db = newDatabase
	return nil
}

// Insert inserts a new list of entries in the database
//
// Return an error if the database can not be initialized
func Insert(actions []*ActionEntry) error {

	if db == nil {
		return errors.New("the database was not created prior to calling Insert")
	}

	// Create a write transaction
	txn := db.Txn(true)

	for _, action := range actions {
		if err := txn.Insert("action", action); err != nil {
			txn.Abort()
			return err
		}
	}

	// Commit the transaction
	txn.Commit()

	return nil
}

// RequestID is the identifier of an HTTP request made during the test. All the same requests done during multiple Test
// or if the Stage is run another time after a failure have the same id.
type RequestID struct {
	StageIndex  int
	ActionIndex int
}

// RequestResults holds the results of each individual HTTP request (action) for a single RequestID
type RequestResults struct {
	NbFailure        int
	SuccessNanoTimes []int
}

// GetAllRequests returns all the HTTP requests executed for a test and their results
//
// Return the requests or an error if something went wrong
func GetAllRequests() (map[RequestID]*RequestResults, error) {

	// Create read-only transaction
	txn := db.Txn(false)
	defer txn.Abort()

	// Get an iterator over all actions
	it, err := txn.Get("action", "id")
	if err != nil {
		return nil, err
	}

	result := make(map[RequestID]*RequestResults)

	// Iterate over all actions
	for obj := it.Next(); obj != nil; obj = it.Next() {
		action := obj.(*ActionEntry)

		requestID := RequestID{
			StageIndex:  action.StageIndex,
			ActionIndex: action.ActionIndex,
		}

		if requestResult, ok := result[requestID]; ok {

			if action.Success {
				requestResult.SuccessNanoTimes = append(requestResult.SuccessNanoTimes, action.DurationNano)
			} else {
				requestResult.NbFailure++
			}
		} else {
			if action.Success {
				result[requestID] = &RequestResults{
					NbFailure: 0,
					SuccessNanoTimes: []int{
						action.DurationNano,
					},
				}
			} else {
				result[requestID] = &RequestResults{
					NbFailure:        1,
					SuccessNanoTimes: make([]int, 0, 10),
				}
			}
		}

	}

	return result, nil
}
