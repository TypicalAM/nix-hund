package db

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// QueryPkg the database using a parameter. The parameter may be in the following formats:
// - "/usr/lib/libc.so.6"
// - "/usr/lib/libc.so"
// - "libc.so.6"
// - "libc.so"
// - "libc"
// Returns a list of resulting packages.
func (db DB) QueryPkg(param string) ([]PkgResult, error) {
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

		rows, err := db.db.Query(query, param)
		if err != nil {
			return nil, err
		}

		return rowsToResult(rows)
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

		rows, err := db.db.Query(query, stripped)
		if err != nil {
			return nil, err
		}

		return rowsToResult(rows)
	}

	query := ""
	if fullPath {
		query = baseQuery + " WHERE full_path = $1"
	} else {
		query = baseQuery + " WHERE filename = $1"
	}

	rows, err := db.db.Query(query, param)
	if err != nil {
		return nil, err
	}

	return rowsToResult(rows)
}

// InsertPkg puts the package information into the index.
func (db *DB) InsertPkg(startTime time.Time, name string, out string, version string, files []string) error {
	const query = "INSERT INTO listings (start_date, pkg_name, output_name, version, path, fullpath, filename) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	for _, path := range files {
		split := strings.Split(path, "/")
		filenameRaw := split[len(split)-1]
		if _, err := tx.Exec(query, startTime, name, out, version, path, path, filenameRaw); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// rowsToResult converts the rows into a []PkgResult object.
func rowsToResult(rows *sql.Rows) ([]PkgResult, error) {
	pkgs := make([]PkgResult, 0)
	for rows.Next() {
		result := PkgResult{}
		if err := rows.Scan(&result.PkgName, &result.Outname, &result.Version, &result.Path); err != nil {
			return nil, err
		}

		pkgs = append(pkgs, result)
	}

	return pkgs, nil
}
