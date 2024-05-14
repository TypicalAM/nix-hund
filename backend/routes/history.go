package routes

import (
	"net/http"

	"github.com/TypicalAM/nix-hund/metrics"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// HistoryList returns the history of the user.
func (cntr *Controller) HistoryList(c echo.Context) error {
	metrics.RequestCount.Inc()

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*JwtUserClaims)
	list, err := cntr.dbase.History(claims.Name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, list)
}

// HistoryDeleteInput specifies which entry should be deleted.
type HistoryDeleteInput struct {
	Index int `json:"idx"`
}

// HistoryDelete deletes the history entry of the user.
func (cntr *Controller) HistoryDelete(c echo.Context) error {
	metrics.RequestCount.Inc()

	input := HistoryDeleteInput{}
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No index supplied")
	}

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*JwtUserClaims)
	if err := cntr.dbase.HistoryDelete(claims.Name, input.Index); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't delete index: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "History entry deleted successfully"})
}
