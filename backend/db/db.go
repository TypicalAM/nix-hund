package db

// Result is a query result, it matches the query to the response.
type Result struct {
	PkgName string
	Outname string
	Path    string
	Version string
}

// IndexDB is an interface that an index should have.
type IndexDB interface {
	// Query the database using a parameter. The parameter may be in the following formats:
	// - "/usr/lib/libc.so.6"
	// - "/usr/lib/libc.so"
	// - "libc.so.6"
	// - "libc.so"
	// - "libc"
	// Returns a list of resulting packages.
	Query(param string) ([]Result, error)

	// Put puts the package information into the index.
	Put(name string, out string, version string, files []string) error
}
