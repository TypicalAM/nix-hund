package routes

import (
	"github.com/TypicalAM/nix-hund/db"
)

// Controller manages the routes.
type Controller struct {
	cacheURL string
	dbase    *db.DB
	channels []string
	cacheDir string
}

// New creates a new controller.
func New(cacheURL string, database *db.DB, channels []string, cacheDir string) (*Controller, error) {
	return &Controller{
		cacheURL: cacheURL,
		dbase:    database,
		channels: channels,
		cacheDir: cacheDir,
	}, nil
}
