package routes

import (
	"net/http"
	"time"

	"github.com/TypicalAM/nix-hund/db"
	"github.com/TypicalAM/nix-hund/metrics"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// RegsterInfo is the information a user gives when signing up.
type registerInfo struct {
	Username string
	Password string
}

// JwtCustomClaims provides a way to save the username.
type JwtUserClaims struct {
	Name string `json:"name"`
	jwt.RegisteredClaims
}

// Register registers a user.
func (cntr *Controller) Register(c echo.Context) error {
	metrics.RequestCount.Inc()
	metrics.RegisterAttempts.Inc()

	var user registerInfo
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err := cntr.dbase.CreateUser(user.Username, user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating user: "+err.Error())
	}

	token, err := createToken(user.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// Login logs the user in.
func (cntr *Controller) Login(c echo.Context) error {
	metrics.RequestCount.Inc()
	metrics.LoginAttempts.Inc()

	testUser := db.User{}
	if err := c.Bind(&testUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := cntr.dbase.QueryUser(testUser.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Querying user failed: "+err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(testUser.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Wrong password")
	}

	token, err := createToken(user.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// createToken creates a JWT token for the user.
func createToken(name string) (string, error) {
	claims := &JwtUserClaims{name, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72))}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret3"))
	return t, err
}

// DeleteUser deletes an account.
func (cntr *Controller) DeleteUser(c echo.Context) error {
	metrics.RequestCount.Inc()

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*JwtUserClaims)
	if err := cntr.dbase.DeleteUser(claims.Name); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted successfully"})
}
