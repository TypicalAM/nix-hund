package main

import (
	"github.com/TypicalAM/nix-hund/db"
	"github.com/TypicalAM/nix-hund/nixpkgs"
	"github.com/TypicalAM/nix-hund/routes"
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

	database, err := db.New()
	if err != nil {
		log.Fatal("Loading db failed", "err", err)
	}

	cntr, err := routes.New(pkgs, database)
	if err != nil {
		log.Fatal("Creating controller failed", "err", err)
	}

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

	e.POST("/register", cntr.Register)
	e.POST("/login", cntr.Login)

	indexGroup := e.Group("/pkg")
	indexGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte("secret"),
		TokenLookup: "header:x-auth-token",
	}))
	indexGroup.GET("/index/generate", cntr.IndexGen)
	indexGroup.GET("/query", cntr.Query)
	indexGroup.GET("/index/list", cntr.IndexList)

	log.Fatal(e.Start(":1323"))
}
