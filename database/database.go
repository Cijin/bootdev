package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id       int    `json:"id,omitempty"`
	AuthorId int    `json:"author_id,omitempty"`
	Body     string `json:"body,omitempty"`
}

type User struct {
	Id           int    `json:"id,omitempty"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	PasswordHash []byte `json:"password_hash,omitempty"`
}

type DbStructure struct {
	Chirps        map[int]Chirp        `json:"chirps,omitempty"`
	Users         map[int]User         `json:"users,omitempty"`
	RevokedTokens map[string]time.Time `json:"revoked_tokens,omitempty"`
}

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("email exists")
	ErrUnAuthorized   = errors.New("unauthorized")
	dbInstance        = &DB{
		"db.json",
		&sync.RWMutex{},
	}
)

// NewDb creates a new database connection
// and creates the database file if it doesn't exist
func NewDb() error {
	err := dbInstance.ensureDB()
	if err != nil {
		return err
	}

	return nil
}

func GetDb() *DB {
	return dbInstance
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(authorId int, body string) (Chirp, error) {
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
		Id:       id,
		AuthorId: authorId,
		Body:     body,
	}
	dbStruct.Chirps[id] = chirp

	err = db.writeDB(dbStruct)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
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

func (db *DB) GetChirp(id int) (Chirp, error) {
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

func (db *DB) DeleteChirp(id int) (Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	c, ok := ds.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotFound
	}

	delete(ds.Chirps, id)

	err = db.writeDB(ds)
	if err != nil {
		return Chirp{}, err
	}

	return c, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	_, ok := db.search(email)
	if ok {
		return User{}, ErrDuplicateEmail
	}

	id := len(dbStruct.Users) + 1
	// nil map
	if id-1 == 0 {
		dbStruct.Users = map[int]User{}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	u := User{
		Id:           id,
		Email:        email,
		PasswordHash: hash,
		IsChirpyRed:  false,
	}
	dbStruct.Users[id] = u

	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return User{Id: u.Id, Email: u.Email, IsChirpyRed: u.IsChirpyRed}, nil
}

func (db *DB) UpdateUser(id int, email string, password string, isChirpyRed bool) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	u, ok := dbStruct.Users[id]
	if !ok {
		return User{}, ErrNotFound
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}

		u.PasswordHash = hash
	}

	if email != "" {
		u.Email = email
	}

	if isChirpyRed {
		u.IsChirpyRed = true
	}

	dbStruct.Users[id] = u

	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return User{Id: u.Id, Email: u.Email, IsChirpyRed: u.IsChirpyRed}, nil
}

func (db *DB) Login(email string, password string) (User, error) {
	u, ok := db.search(email)
	if !ok {
		return User{}, ErrNotFound
	}

	err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
	if err != nil {
		return User{}, ErrUnAuthorized
	}

	return User{Id: u.Id, Email: u.Email, IsChirpyRed: u.IsChirpyRed}, nil
}

func (db *DB) RevokeToken(token string) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	size := len(dbStruct.RevokedTokens)
	// nil map
	if size == 0 {
		dbStruct.RevokedTokens = map[string]time.Time{}
	}

	dbStruct.RevokedTokens[token] = time.Now()

	return db.writeDB(dbStruct)
}

func (db *DB) IsRevoked(token string) (bool, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return false, err
	}

	_, ok := dbStruct.RevokedTokens[token]

	return ok, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		_, err = os.Create(db.path)
	}

	if err != nil {
		return err
	}

	dbStruct := DbStructure{make(map[int]Chirp), make(map[int]User), make(map[string]time.Time)}
	return db.writeDB(dbStruct)
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DbStructure, error) {
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
func (db *DB) writeDB(ds DbStructure) error {
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

// search
func (db *DB) search(email string) (User, bool) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, false
	}

	for _, u := range dbStruct.Users {
		if u.Email == email {
			return u, true
		}
	}

	return User{}, false
}
