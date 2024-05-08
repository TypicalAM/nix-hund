package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"

	"github.com/TypicalAM/nix-hund/db"
	"github.com/TypicalAM/nix-hund/nixpkgs"
	"github.com/labstack/echo/v4"
)

// Controller manages the routes.
type Controller struct {
	pkgs  *nixpkgs.Pkgs
	index db.IndexDB
}

// New creates a new controller.
func New(pkgs *nixpkgs.Pkgs, index db.IndexDB) (*Controller, error) {
	return &Controller{
		pkgs:  pkgs,
		index: index,
	}, nil
}

// Index creates an index.
func (cntr *Controller) Index(c echo.Context) error {
	totalFileCount := 0
	totalPkgs := 0
	start := time.Now()

	for listing := range cntr.pkgs.ProcessListings(cntr.pkgs.FetchListings(cntr.pkgs.CountDev())) {
		if err := cntr.index.Put(listing.PkgName, listing.OutputName, "", listing.Files); err != nil {
			log.Fatal("Indexing failed", "name", listing.PkgName, "err", err)
		}

		totalFileCount += len(listing.Files)
		totalPkgs++

		log.Info("Package",
			"name", listing.PkgName,
			"outname", listing.OutputName,
			"size", len(listing.Files),
			"total_packages", totalPkgs,
			"total_files", totalFileCount,
		)
	}

	log.Info("Indexing done", "time took", time.Now().Sub(start))
	return c.String(http.StatusOK, "Indexing done")
}
