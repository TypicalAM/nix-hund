package routes

import (
	"github.com/TypicalAM/nix-hund/metrics"
	"net/http"
	"time"

	"github.com/TypicalAM/nix-hund/db"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Username string
	Password string
}

// Register registers a user.
func (cntr *Controller) Register(c echo.Context) error {
	metrics.RequestCount.Inc()
	metrics.RegisterAttempts.Inc()

	var user RegisterUser
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err := cntr.database.CreateUser(user.Username, user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating user: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Registered successfully"})
}

// createToken creates a JWT token for the user.
func createToken(name string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = name
	claims["exp"] = time.Now().Add(time.Hour * 24 * 10).Unix()
	t, err := token.SignedString([]byte("secret"))
	return t, err
}

// Login logs the user in.
func (cntr *Controller) Login(c echo.Context) error {
	metrics.RequestCount.Inc()
	metrics.LoginAttempts.Inc()

	testUser := db.User{}
	if err := c.Bind(&testUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := cntr.database.QueryUser(testUser.Username)
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
