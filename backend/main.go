package main

import (
	"net/http"

	"github.com/TypicalAM/nix-hund/nixpkgs"
	"github.com/charmbracelet/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const CACHE_URL = "http://cache.nixos.org"

func main() {
	pkgs, err := nixpkgs.New(CACHE_URL)
	if err != nil {
		log.Fatal("Loading packages failed", "err", err)
	}

	log.Info("Discovering ended", "pkgs", pkgs.Count())

	if err = pkgs.CreateIndex(); err != nil {
		log.Fatal("Creating index failed", "err", err)
	}

	log.Info("Starting router")

	return
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Debug("URI", v.URI, "status", v.Status)
			return nil
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	log.Fatal(e.Start(":1323"))
}
