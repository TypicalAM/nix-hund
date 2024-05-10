package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
)

// IndexGen creates an index.
func (cntr *Controller) IndexGen(c echo.Context) error {
	startTime := time.Now()
	totalFileCount := 0
	totalPkgs := 0
	start := time.Now()

	for listing := range cntr.pkgs.ProcessListings(cntr.pkgs.FetchListings(cntr.pkgs.CountDev())) {
		if err := cntr.database.InsertPkg(startTime, listing.PkgName, listing.OutputName, "", listing.Files); err != nil {
			log.Error("Indexing failed", "name", listing.PkgName, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Indexing failed at: "+listing.PkgName)
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
	return c.JSON(http.StatusOK, map[string]string{"message": "Indexing done"})
}

// IndexList lists the available indices.
func (cntr *Controller) IndexList(c echo.Context) error {
	indices, err := cntr.database.ListIndices()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error finding indices: "+err.Error())
	}

	return c.JSON(http.StatusOK, indices)
}

// Query queries for a package.
func (cntr *Controller) Query(c echo.Context) error {
	param := c.QueryParam("query")
	if param == "" {
		return c.String(http.StatusBadRequest, "No query param")
	}

	res, err := cntr.database.QueryPkg(param)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error while fetching")
	}

	data, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error while marshaling")
	}

	return c.String(http.StatusOK, string(data))
}
