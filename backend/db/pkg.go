package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// PkgResult is a package query result from the index.
type PkgResult struct {
	PkgName string `json:"pkg_name"`
	Outname string `json:"out_name"`
	Outhash string `json:"out_hash"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

// IndexInfo is the short information about a previously created index, if the index was short lived, it should not be used.
type IndexInfo struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	FileCount int       `json:"total_file_count"`
}

// ListIndices lists the available indices in the database.
func (db *DB) ListIndices(channel string) ([]IndexInfo, error) {
	rows, err := db.db.Query(`SELECT index_uuid, index_date, COUNT(*) FROM listings WHERE index_channel = $1 GROUP BY index_date ORDER BY index_date DESC`, channel)
	if err != nil {
		return nil, fmt.Errorf("listing indices: %w", err)
	}
	defer rows.Close()

	indices := make([]IndexInfo, 0)
	for rows.Next() {
		index := IndexInfo{}
		if err := rows.Scan(&index.ID, &index.Date, &index.FileCount); err != nil {
			return nil, fmt.Errorf("scanning indices rows: %w", err)
		}
		indices = append(indices, index)
	}

	return indices, nil
}

// QueryPkg the database using a parameter. The parameter may be in the following formats:
// - "/usr/lib/libc.so.6"
// - "libc.so.6"
// Returns a list of resulting packages.
func (db DB) QueryPkg(id, param string) ([]PkgResult, error) {
	const baseQuery = "SELECT pkg_name, output_name, output_hash, version, fullpath FROM listings"
	log.Info("Querying index", "param", param)
	fullPath := strings.Count(param, "/") > 1
	query := ""

	if fullPath {
		query = baseQuery + " WHERE fullpath = $1 AND index_uuid = $2"
	} else {
		query = baseQuery + " WHERE filename = $1 AND index_uuid = $2"
	}

	rows, err := db.db.Query(query, param, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToResult(rows)
}

// InsertPkg puts the package information into the index.
func (db *DB) InsertPkg(indexDate time.Time, channel, id, name, out, hash, version string, files []string) error {
	const query = `INSERT INTO listings (index_channel, index_date, index_uuid, pkg_name, output_name,	output_hash, version, fullpath, filename)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	for _, path := range files {
		split := strings.Split(path, "/")
		filename := split[len(split)-1]
		if _, err := tx.Exec(query, channel, indexDate, id, name, out, hash, version, path, filename); err != nil {
			fmt.Println(indexDate, id, name, out, hash, version, path, filename)
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
		if err := rows.Scan(&result.PkgName, &result.Outname, &result.Outhash, &result.Version, &result.Path); err != nil {
			return nil, err
		}

		pkgs = append(pkgs, result)
	}

	return pkgs, nil
}
