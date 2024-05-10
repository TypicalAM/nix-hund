package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

// PkgResult is a package query result from the index.
type PkgResult struct {
	PkgName string `json:"pkg_name"`
	Outname string `json:"out_name"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

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

// IndexInfo is the short information about a previously created index, if the index was short lived, it should not be used.
type IndexInfo struct {
	Date        time.Time
	OutputCount int
}

// ListIndices lists the available indices in the database.
func (db *DB) ListIndices() ([]IndexInfo, error) {
	rows, err := db.db.Query(`SELECT start_date, COUNT(*) FROM listings GROUP BY start_date;`)
	if err != nil {
		return nil, fmt.Errorf("listing indices: %w", err)
	}

	indices := make([]IndexInfo, 0)
	for rows.Next() {
		index := IndexInfo{}
		if err := rows.Scan(&index.Date, &index.OutputCount); err != nil {
			return nil, fmt.Errorf("scanning indices rows: %w", err)
		}
		indices = append(indices, index)
	}

	return indices, nil
}

// initialize initializes the database fields.
func initialize(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS listings (
		start_date DATE NOT NULL, 
		pkg_name VARCHAR(255) NOT NULL,
		output_name VARCHAR(255) NOT NULL,
		version VARCHAR(50) NOT NULL,
		path VARCHAR(255) NOT NULL,
		fullpath VARCHAR(255) NOT NULL,
		filename VARCHAR(255) NOT NULL,
		PRIMARY KEY (start_date, fullpath, version, pkg_name, output_name)
	);

	CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY NOT NULL,
    password VARCHAR(255) NOT NULL
	);
	`)

	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}

	return nil
}
