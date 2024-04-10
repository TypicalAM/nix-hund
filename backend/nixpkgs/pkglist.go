package nixpkgs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	retryhttp "github.com/hashicorp/go-retryablehttp"
)

const APP_NAME = "nix-hund"

var ErrNoCache = errors.New("no adequate cache")

// info is the information about an available nixpkgs derivation.
type info struct {
	Name       string               `json:"name"`
	OutputName string               `json:"outputName"`
	Outputs    map[string]StorePath `json:"outputs"`
	Pname      string               `json:"pname"`
	System     string               `json:"system"`
	Version    string               `json:"version"`
}

// list is the nixpkgs package list.
type list map[string]info

// Pkgs is the package list fetcher
type Pkgs struct {
	CacheURL string
	List     list
	Fetcher  *retryhttp.Client
}

// New reads or fetches the available packages from nixpkgs. It uses the default system channel and the cache url provided by the caller.
func New(url string) (*Pkgs, error) {
	cli := retryhttp.NewClient()
	cli.RetryMax = 5
	cli.Backoff = retryhttp.LinearJitterBackoff
	cli.Logger = log.New(io.Discard)

	list, err := listFromCache()
	if err == nil {
		return &Pkgs{
			CacheURL: url,
			List:     list,
			Fetcher:  cli,
		}, nil
	}

	list, err = fetchList()
	if err != nil {
		return nil, err
	}

	return &Pkgs{
		CacheURL: url,
		List:     list,
		Fetcher:  cli,
	}, nil
}

// Count returns the total number of all derivations (NOT all outputs).
func (pkgs Pkgs) Count() int {
	return len(pkgs.List)
}

// RawListing is the information about a listing.
type RawListing struct {
	PkgName    string
	OutputName string
	Data       []byte
	Count      int
}

// Listing is a listing broken down into individual files.
type Listing struct {
	PkgName    string
	OutputName string
	Files      []string
}

// CountDev counts the development packages.
func (pkgs *Pkgs) CountDev() int {
	total := 0

	for _, pkg := range pkgs.List {
		for outname := range pkg.Outputs {
			if outname == "dev" {
				total++
			}
		}
	}

	return total
}

// FetchListings fetches listings for all files.
func (pkgs *Pkgs) FetchListings(total int) chan RawListing {
	doneCount := atomic.Int64{}
	result := make(chan RawListing)
	count := 0

	for pkgName, pkg := range pkgs.List {
		for outname, sp := range pkg.Outputs {
			if outname != "dev" {
				continue
			}

			count++
			go pkgs.fetchPackage(pkgName, sp, outname, total, &doneCount, result)
		}

		if count >= total {
			break
		}
	}

	return result
}

// fetchPackage fetches a raw file listing.
func (pkgs *Pkgs) fetchPackage(pkgName string, sp StorePath, outname string, total int, doneCount *atomic.Int64, listings chan RawListing) {
	data, err := sp.FetchListing(pkgs.CacheURL, pkgs.Fetcher.StandardClient())
	doneCount.Store(doneCount.Add(1))
	if err != nil {
		if doneCount.Load() == int64(total) {
			log.Info("Done fetching", "pkgs", total)
			close(listings)
		}

		log.Error("Failed to fetch listing", "name", pkgName, "err", err)
		return
	}

	listings <- RawListing{
		PkgName:    pkgName,
		OutputName: outname,
		Data:       data,
		Count:      int(doneCount.Load()),
	}

	if doneCount.Load() == int64(total) {
		log.Info("Done fetching", "pkgs", total)
		close(listings)
	}
}

// ProcessListings processes a channel of packages into resolved file listings.
func (pkgs *Pkgs) ProcessListings(rawPkgs chan RawListing) chan Listing {
	result := make(chan Listing)

	go func() {
		for raw := range rawPkgs {
			go pkgs.processInfo(raw, result)
		}
		close(result)
	}()

	return result
}

// processInfo resolves the raw listing info to a filelist, and saves it to disk.
func (pkgs *Pkgs) processInfo(raw RawListing, listings chan Listing) {
	files, err := GetFileList(raw.Data)
	if err != nil {
		log.Error("Failed to fetch listing", "name", raw.PkgName, "err", err)
		return
	}

	listings <- Listing{
		PkgName:    raw.PkgName,
		OutputName: raw.OutputName,
		Files:      files,
	}
}

// listFromCache gets the package list from the user's cache directory.
func listFromCache() (list, error) {
	log.Debug("Trying to find cache")
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := dir + "/" + APP_NAME + "/lists"
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err2 := os.MkdirAll(path, 0666); err2 != nil {
			return nil, err2
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dates := make([]time.Time, 0, len(entries))
	for _, entry := range entries {
		// Expected name pattern "2006-01-02 15:04:05.json"
		name, found := strings.CutSuffix(entry.Name(), ".json")
		if !found {
			return nil, errors.New("invalid file in cache")
		}

		date, err := time.Parse(time.DateTime, name)
		if err != nil {
			return nil, err
		}

		dates = append(dates, date)
	}

	if len(dates) == 0 {
		return nil, ErrNoCache
	}

	maxDate := slices.MaxFunc(dates, func(a time.Time, b time.Time) int {
		if a.After(b) {
			return 1
		} else if a.Equal(b) {
			return 0
		}
		return -1
	})

	idx := slices.Index(dates, maxDate)
	if dates[idx].Before(time.Now().Add(-24 * time.Hour)) {
		return nil, ErrNoCache
	}

	latestPath := path + "/" + entries[idx].Name()
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, err
	}

	result := list{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// fetchList fetches the package list from the remote server.
func fetchList() (list, error) {
	log.Info("Fetching pkg lists")

	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := dir + "/" + APP_NAME + "/lists"
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		if !os.IsNotExist(err) {
			return nil, err
		}

		fmt.Println(path)
		if err2 := os.Mkdir(path, 0755); err2 != nil {
			return nil, err2
		}
	}

	filename := path + "/" + time.Now().Format(time.DateTime) + ".json"
	outfile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer outfile.Close()

	buf := &bytes.Buffer{}
	cmd := exec.Command(
		"nix-env",
		"--out-path",
		"--query",
		"--available",
		"--json",
		"--arg", "config", "{ allowAliases = false; }",
		"--argstr", "system", "x86_64-linux",
		"--prebuilt-only",
		"--show-trace",
	)
	cmd.Stdout = io.MultiWriter(outfile, buf)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := list{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}

	log.Info("Fetching done", "pkgs", len(result))
	return result, nil
}
