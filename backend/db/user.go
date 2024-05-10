package db

import (
	"errors"

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
	res, err := db.db.Exec(`DELETE FROM users WHERE username = $1`, username, username)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("Insert failed, no rows affected")
	}

	return nil
}
