package routes

import (
	"github.com/TypicalAM/nix-hund/db"
	"github.com/TypicalAM/nix-hund/nixpkgs"
)

// Controller manages the routes.
type Controller struct {
	pkgs     *nixpkgs.Pkgs
	database *db.DB
}

// New creates a new controller.
func New(pkgs *nixpkgs.Pkgs, database *db.DB) (*Controller, error) {
	return &Controller{
		pkgs:     pkgs,
		database: database,
	}, nil
}
