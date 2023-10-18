package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Db struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id,omitempty"`
	Body string `json:"body,omitempty"`
}

type User struct {
	Id           int    `json:"id,omitempty"`
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	PasswordHash []byte `json:"password_hash,omitempty"`
}

type DbStructure struct {
	Chirps map[int]Chirp   `json:"chirps,omitempty"`
	Users  map[string]User `json:"users,omitempty"`
}

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("email exists")
	ErrUnAuthorized   = errors.New("unauthorized")
)

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

func (db *Db) GetChirp(id int) (Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	c, ok := ds.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotFound
	}

	return c, nil
}

func (db *Db) CreateUser(email string, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if _, ok := dbStruct.Users[email]; ok {
		return User{}, ErrDuplicateEmail
	}

	id := len(dbStruct.Users) + 1
	// nil map
	if id-1 == 0 {
		dbStruct.Users = map[string]User{}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	u := User{
		Id:           id,
		Email:        email,
		PasswordHash: hash,
	}
	dbStruct.Users[email] = u

	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return User{Id: u.Id, Email: u.Email}, nil
}

func (db *Db) Login(email string, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	u, ok := dbStruct.Users[email]
	if !ok {
		return User{}, ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
	if err != nil {
		return User{}, ErrUnAuthorized
	}

	return User{Id: u.Id, Email: u.Email}, nil
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

	dbStruct := DbStructure{make(map[int]Chirp), make(map[string]User)}
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
