package main

import (
	"flag"
	"os"

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
var fetchChannel = flag.String("fetch", "", "Channel name for fetching data, something that can be put in `nix-env --file`. For example `channel:nixos-21.11` or a nixpkgs archive url, should be paired with --out_path")
var outpath = flag.String("out_path", "", "Output path for a dumped channel. ~/.cache/nix-hund/channels/file.json is appropriate for reading by the program")
var cacheDir = flag.String("cache_dir", "", "Cache directory to use instead of the default one")

const CACHE_URL = "http://cache.nixos.org"

func main() {
	flag.Parse()

	if *fetchChannel != "" {
		if *outpath == "" {
			log.Fatal("Specified the fetch channel without an out_path, use --out_path to tell nix-hund where to put the result of the fetch")
		}

		if err := nixpkgs.FetchChannel(*fetchChannel, *outpath); err != nil {
			log.Fatal("Cannot fetch the data for a channel", "err", err)
		}

		return
	}

	cache := *cacheDir
	if cache == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			log.Fatal("Cannot get the cache directory", "err", err)
		}
		cache = userCache
	}

	dir := cache + "/nix-hund"
	stat, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatal("Couldn't create directory", "err", err)
			}

			stat, err = os.Stat(dir)
			if err != nil {
				log.Fatal("Cannot stat the cache directory", "err", err)
			}
		} else {
			log.Fatal("Cannot stat the cache directory", "err", err)
		}
	}

	if !stat.IsDir() {
		log.Fatal("Cache directoy exists and isn't a directory", "dir", dir)
	}

	database, err := db.New(*cacheDir)
	if err != nil {
		log.Fatal("Loading db failed", "err", err)
	}

	channels, err := nixpkgs.AvailableChannels(*cacheDir)
	if err != nil {
		log.Fatal("Couldn't get the available channels", "err", err)
	}

	cntr, err := routes.New(CACHE_URL, database, channels, *cacheDir)
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

	protected := echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(routes.JwtUserClaims) },
		SigningKey:    []byte(os.Getenv("HUND_SECRET_KEY")),
	})

	accounts := e.Group("/account")
	accounts.POST("/register", cntr.Register)
	accounts.POST("/login", cntr.Login)
	accounts.GET("/history", cntr.HistoryList, protected)
	accounts.POST("/history/delete", cntr.HistoryDelete, protected)
	accounts.POST("/delete", cntr.DeleteUser, protected)

	pkgs := e.Group("/pkg")
	pkgs.GET("/channel", cntr.ChannelList)
	pkgs.GET("/channel/index", cntr.IndexList)
	pkgs.POST("/channel/index/generate", cntr.IndexGenerate, protected)
	pkgs.GET("/index/:id/query", cntr.IndexQuery, protected)

	if *disableMetrics {
		e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	}

	log.Fatal(e.Start(":1323"))
}
