package db

import (
	"errors"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/crypto/bcrypt"
)

var ErrNoUser = errors.New("no user with this username")

// User is a basic user.
type User struct {
	Username string
	Password string
}

// QueryUser returns a user based on the username.
func (db *DB) QueryUser(username string) (*User, error) {
	rows, err := db.db.Query(`SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user User
	if !rows.Next() {
		return nil, ErrNoUser
	}

	if err := rows.Scan(&user.Username, &user.Password); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a user in the database.
func (db *DB) CreateUser(username, password string) (*User, error) {
	_, err := db.QueryUser(username)
	if err == nil {
		return nil, errors.New("already exists")
	}

	if len(password) < 8 {
		return nil, errors.New("password too short")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // TODO: Salting
	if err != nil {
		return nil, err
	}

	res, err := db.db.Exec(`INSERT INTO users VALUES ($1, $2)`, username, string(hashed))
	if err != nil {
		return nil, err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("Insert failed, no rows affected")
	}

	return &User{Username: username, Password: password}, nil
}

// DeleteUser deletes a user from the database.
func (db *DB) DeleteUser(username string) error {
	res, err := db.db.Exec(`DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("Insert failed, no rows affected")
	}

	return nil
}

// HistoryEntry is an entry in the user's search history.
type HistoryEntry struct {
	IndexID string    `json:"index_id"`
	Date    time.Time `json:"date"`
	Pkg     PkgResult `json:"pkg"`
}

// History returns the package search history of the user.
func (db *DB) History(username string) ([]HistoryEntry, error) {
	const query = `SELECT index_uuid, date, pkg_name, output_name, output_hash, fullpath, version
FROM users_history NATURAL JOIN listings WHERE username = $1 ORDER BY date DESC`
	rows, err := db.db.Query(query, username)
	if err != nil {
		log.Error("Error while getting history", "err", err)
		return nil, err
	}
	defer rows.Close()

	history := make([]HistoryEntry, 0)
	for rows.Next() {
		entry := HistoryEntry{}
		if err := rows.Scan(
			&entry.IndexID, &entry.Date, &entry.Pkg.PkgName, &entry.Pkg.Outname, &entry.Pkg.Outhash, &entry.Pkg.Path, &entry.Pkg.Version,
		); err != nil {
			log.Error("Error while scanning history", "err", err)
			return nil, err
		}
		history = append(history, entry)
	}

	return history, nil
}

// HistoryAdd adds a history entry for a user.
func (db *DB) HistoryAdd(id string, username string, result PkgResult) error {
	const query = `INSERT INTO users_history values ($1, $2, $3, $4, $5, $6)`
	res, err := db.db.Exec(query, username, time.Now(), result.PkgName, id, result.Outhash, result.Path)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("Insert failed, no rows affected")
	}

	log.Info("Added a history entry", "user", username, "pkgname", result.PkgName)
	return nil
}

// HistoryDelete deletes a history entry by index.
func (db *DB) HistoryDelete(username string, idx int) error {
	const selection = `SELECT date FROM users_history WHERE username = $1 ORDER BY date DESC limit 1 offset $2`
	rows, err := db.db.Query(selection, username, idx)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return errors.New("No index like this one")
	}

	date := time.Time{}
	if err := rows.Scan(&date); err != nil {
		return err
	}
	rows.Close()

	const deletion = `DELETE FROM users_history WHERE username = $1 AND date = $2`
	res, err := db.db.Exec(deletion, username, date)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("Delete failed, no rows affected")
	}

	log.Info("Deleted history entry", "user", username, "idx", idx)
	return nil
}
