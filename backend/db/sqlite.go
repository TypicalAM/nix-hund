package db

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Database is the index database.
type SqliteIndex struct {
	db  *sql.DB
	url string
}

// New creates a new sqlite database in the cache path.
func NewSqlite() (IndexDB, error) {
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

	return SqliteIndex{
		db:  db,
		url: path,
	}, nil
}

// initialize initializes the database fields.
func initialize(db *sql.DB) error {
	db.Exec(`CREATE TABLE listings (
    pkg_name VARCHAR(255),
    output_name VARCHAR(255),
    version VARCHAR(50),
    path VARCHAR(255),
    fullpath VARCHAR(255),
    filename VARCHAR(255)
		);`)

	return nil
}

// Query fetches the query parameter from the sqlite database.
func (idx SqliteIndex) Query(param string) ([]Result, error) {
	const baseQuery = "SELECT pkg_name, output_name, version, path FROM listings"
	fullPath := strings.Count(param, "/") > 1
	hasDots := strings.Count(param, ".") > 0

	if !hasDots {
		// Probably an executable
		query := ""
		if fullPath {
			query = baseQuery + " WHERE full_path = $1"
		} else {
			query = baseQuery + " WHERE filename = $1"
		}

		rows, err := idx.db.Query(query, param)
		if err != nil {
			return nil, err
		}

		return convertResults(rows), nil
	}

	pointIdx := strings.IndexRune(param, '.')
	multipleDots := strings.Count(param, ".") > 1
	if multipleDots {
		// Cut off the version
		nextPointIdx := strings.IndexRune(param[pointIdx:], '.')
		stripped := param[:pointIdx] + param[pointIdx:nextPointIdx]
		query := ""
		if fullPath {
			query = baseQuery + " WHERE full_path = $1"
		} else {
			query = baseQuery + " WHERE filename = $1"
		}

		rows, err := idx.db.Query(query, stripped)
		if err != nil {
			return nil, err
		}

		return convertResults(rows), nil
	}

	query := ""
	if fullPath {
		query = baseQuery + " WHERE full_path = $1"
	} else {
		query = baseQuery + " WHERE filename = $1"
	}

	rows, err := idx.db.Query(query, param)
	if err != nil {
		return nil, err
	}

	return convertResults(rows), nil
}

// Put updates the value in the database.
func (idx SqliteIndex) Put(name string, out string, version string, files []string) error {
	const query = "INSERT INTO listings (pkg_name, output_name, version, path, fullpath, filename) VALUES ($1, $2, $3, $4, $5, $6)"

	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}

	for _, path := range files {
		split := strings.Split(path, "/")
		filenameRaw := split[len(split)-1]
		if _, err := tx.Exec(query, name, out, version, path, path, filenameRaw); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// convertResults converts the rows into a []Result object.
func convertResults(rows *sql.Rows) []Result {
	aggregate := make([]Result, 0)
	for rows.Next() {
		result := Result{}
		if err := rows.Scan(&result.PkgName, &result.Outname, &result.Version, &result.Path); err != nil {
			panic(err)
		}

		aggregate = append(aggregate, result)
	}

	return aggregate
}
