package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database of the program. It handles users and packages.
type DB struct {
	db   *sql.DB
	path string
}

// New creates a new db instance.
func New() (*DB, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := dir + "/nix-hund/index.db"
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := initialize(db); err != nil {
		return nil, err
	}

	return &DB{
		db:   db,
		path: path,
	}, nil
}

// initialize initializes the database fields.
func initialize(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS listings (
		index_channel VARCHAR(255) NOT NULL,
		index_date DATE NOT NULL, 
		index_uuid CHAR(36) NOT NULL,
		pkg_name VARCHAR(255) NOT NULL,
		output_name VARCHAR(255) NOT NULL,
		output_hash VARCHAR(255) NOT NULL,
		version VARCHAR(50) NOT NULL,
		fullpath VARCHAR(255) NOT NULL,
		filename VARCHAR(255) NOT NULL,
		PRIMARY KEY (pkg_name, index_uuid, output_hash, fullpath)
	);

	CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY NOT NULL,
    password VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS users_history (
		username VARCHAR(255) NOT NULL,
		date DATE NOT NULL,
		pkg_name VARCHAR(255) NOT NULL,
		index_uuid CHAR(36) NOT NULL,
		output_hash VARCHAR(255) NOT NULL,
		fullpath VARCHAR(255) NOT NULL,
		FOREIGN KEY (username) REFERENCES users(user_id),
		FOREIGN KEY (pkg_name, index_uuid, output_hash, fullpath) REFERENCES listings(pkg_name, index_uuid, output_hash, fullpath)
	);
	`)

	return err
}
