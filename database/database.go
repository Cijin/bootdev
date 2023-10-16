package database

import (
	"encoding/json"
	"os"
	"sync"
)

type Db struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id,omitempty"`
	Body string `json:"body,omitempty"`
}

type DbStructure struct {
	Chirps map[int]Chirp `json:"chirps,omitempty"`
}

// NewDb creates a new database connection
// and creates the database file if it doesn't exist
func NewDb(path string) (*Db, error) {
	db := &Db{
		path,
		&sync.RWMutex{},
	}

	err := db.ensureDB()
	return db, err
}

// CreateChirp creates a new chirp and saves it to disk
func (db *Db) CreateChirp(body string) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStruct.Chirps) + 1
	// nil map
	if id-1 == 0 {
		dbStruct.Chirps = map[int]Chirp{}
	}

	chirp := Chirp{
		Id:   id,
		Body: body,
	}
	dbStruct.Chirps[id] = chirp

	err = db.writeDB(dbStruct)
	if err != nil {
		delete(dbStruct.Chirps, id)
		id--

		return Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *Db) GetChirps() ([]Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(ds.Chirps))
	for _, v := range ds.Chirps {
		chirps = append(chirps, v)
	}

	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *Db) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		_, err = os.Create(db.path)
	}

	if err != nil {
		return err
	}

	dbStruct := DbStructure{make(map[int]Chirp)}
	return db.writeDB(dbStruct)
}

// loadDB reads the database file into memory
func (db *Db) loadDB() (DbStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	ds := DbStructure{}

	buf, err := os.ReadFile(db.path)
	if err != nil {
		return ds, err
	}

	err = json.Unmarshal(buf, &ds)
	if err != nil {
		return ds, err
	}

	return ds, nil
}

// writeDB writes the database file to disk
func (db *Db) writeDB(ds DbStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	buf, err := json.MarshalIndent(ds, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, buf, 0666)
	if err != nil {
		return err
	}

	return nil
}
