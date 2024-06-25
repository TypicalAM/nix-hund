package routes

import (
	"net/http"
	"slices"
	"time"

	"github.com/TypicalAM/nix-hund/metrics"
	"github.com/TypicalAM/nix-hund/nixpkgs"
	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ChannelList lists the available channels.
type ChannelList struct {
	Channels []string `json:"channels"`
}

// ChannelList lists the available channels.
func (cntr *Controller) ChannelList(c echo.Context) error {
	return c.JSON(http.StatusOK, ChannelList{Channels: cntr.channels})
}

// IndexList returns indices made on this channel.
func (cntr *Controller) IndexList(c echo.Context) error {
	channel := c.QueryParam("channel")
	if channel == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No channel specified")
	}

	list, err := cntr.dbase.ListIndices(channel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error while listing indices: "+err.Error())
	}

	return c.JSON(http.StatusOK, list)
}

// IndexGenerateInput specifies the channel that we want to use.
type IndexGenerateInput struct {
	Channel string `json:"channel"`
}

// IndexGenreateResult is the result information about the freshly generated index.
type IndexGenreateResult struct {
	ID                string        `json:"id"`
	Time              time.Duration `json:"time"`
	TotalPackageCount int           `json:"total_package_count"`
	TotalFilecount    int           `json:"total_file_count"`
}

// IndexGenerate creates an index for a channel.
func (cntr *Controller) IndexGenerate(c echo.Context) error {
	metrics.RequestCount.Inc()
	metrics.IndexCount.Inc()

	input := IndexGenerateInput{}
	if err := c.Bind(&input); err != nil {
		log.Error("Error binding while index generation", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !slices.Contains(cntr.channels, input.Channel) {
		log.Error("Asked for a bad channel", "channel", input.Channel)
		return echo.NewHTTPError(http.StatusBadRequest, "This channel isn't parsed, use /channel to get the available channels")
	}

	pkgs, err := nixpkgs.New(cntr.cacheURL, input.Channel, cntr.cacheDir)
	if err != nil {
		log.Error("Indexing failed", "channel", input.Channel, "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Getting lists failed: "+err.Error())
	}

	indexTime := time.Now()
	totalFileCount := 0
	totalPkgs := 0
	id := uuid.New().String()

	for listing := range pkgs.ProcessListings(pkgs.FetchListings(pkgs.CountDev())) {
		if err := cntr.dbase.InsertPkg(indexTime, input.Channel, id, listing.PkgName, listing.OutputName, listing.OutputHash, listing.Version, listing.Files); err != nil {
			log.Error("Indexing failed", "name", listing.PkgName, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Indexing failed at: "+listing.PkgName)
		}

		totalFileCount += len(listing.Files)
		totalPkgs++

		metrics.ProcessedOutputsCount.Inc()
		log.Info("Package",
			"name", listing.PkgName,
			"outname", listing.OutputName,
			"size", len(listing.Files),
			"total_packages", totalPkgs,
			"total_files", totalFileCount,
		)
	}

	end := time.Now().Sub(indexTime)
	log.Info("Indexing done", "time taken", end)
	return c.JSON(http.StatusOK, IndexGenreateResult{
		ID:                id,
		Time:              end,
		TotalFilecount:    totalFileCount,
		TotalPackageCount: totalPkgs,
	})
}

// Query queries an index for a package.
func (cntr *Controller) IndexQuery(c echo.Context) error {
	metrics.RequestCount.Inc()

	query := c.QueryParam("query")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No query param")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No index id")
	}

	res, err := cntr.dbase.QueryPkg(id, query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error while fetching: "+err.Error())
	}

	if len(res) == 0 {
		return c.JSON(http.StatusOK, res)
	}

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*JwtUserClaims)
	if err := cntr.dbase.HistoryAdd(id, claims.Name, res[0]); err != nil { // TODO: Saving only the first result
		return echo.NewHTTPError(http.StatusInternalServerError, "Error while adding history entry: "+err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
