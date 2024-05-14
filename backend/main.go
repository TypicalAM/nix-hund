package main

import (
	"flag"

	"github.com/TypicalAM/nix-hund/db"
	"github.com/TypicalAM/nix-hund/nixpkgs"
	"github.com/TypicalAM/nix-hund/routes"
	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var disableMetrics = flag.Bool("metrics", true, "Show metrics")
var fetchChannel = flag.String("fetch", "", "Channel name for fetching data, something that can be put in `nix-env --file`. For example `channel:nixos-21.11` or a nixpkgs archive url, should be paired with --outfile")
var outpath = flag.String("out_path", "", "Output path for a dumped channel. ~/.cache/nix-hund/channels/file.json is appropriate for reading by the program")

const CACHE_URL = "http://cache.nixos.org"

func main() {
	flag.Parse()

	if *fetchChannel != "" {
		if *outpath == "" {
			log.Fatal("Specified the fetch channel without an outfile, use --outfile to tell nix-hund where to put the result of the fetch")
		}

		if err := nixpkgs.FetchChannel(*fetchChannel, *outpath); err != nil {
			log.Fatal("Cannot fetch the data for a channel", "err", err)
		}

		return
	}

	database, err := db.New()
	if err != nil {
		log.Fatal("Loading db failed", "err", err)
	}

	channels, err := nixpkgs.AvailableChannels()
	if err != nil {
		log.Fatal("Couldn't get the available channels", "err", err)
	}

	cntr, err := routes.New(CACHE_URL, database, channels)
	if err != nil {
		log.Fatal("Creating controller failed", "err", err)
	}

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("Request", "URI", v.URI, "status", v.Status)
			return nil
		},
	}))

	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(routes.JwtUserClaims) },
		SigningKey:    []byte("secret3"),
	})

	accounts := e.Group("/account")
	accounts.POST("/register", cntr.Register)
	accounts.POST("/login", cntr.Login)
	accounts.GET("/history", cntr.HistoryList, jwtMiddleware)
	accounts.POST("/history/delete", cntr.HistoryDelete, jwtMiddleware)
	accounts.POST("/delete", cntr.DeleteUser, jwtMiddleware)

	pkgs := e.Group("/pkg")
	pkgs.Use(jwtMiddleware)
	pkgs.GET("/channel", cntr.ChannelList)
	pkgs.GET("/channel/index", cntr.IndexList)
	pkgs.POST("/channel/index/generate", cntr.IndexGenerate)
	pkgs.GET("/index/:id/query", cntr.IndexQuery)

	if *disableMetrics {
		e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	}

	log.Fatal(e.Start(":1323"))
}
