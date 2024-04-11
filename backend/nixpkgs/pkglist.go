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
	"sync"
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
	Count      int
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
	wg := sync.WaitGroup{}
	rawListings := make(chan RawListing)
	count := 0

	for pkgName, pkg := range pkgs.List {
		for outname, sp := range pkg.Outputs {
			if outname != "dev" {
				continue
			}

			wg.Add(1)
			count++
			go pkgs.fetchPackage(pkgName, sp, outname, &wg, count, rawListings)
		}

		if count >= total {
			break
		}
	}

	go func() {
		wg.Wait()
		close(rawListings)
	}()

	return rawListings
}

// fetchPackage fetches a raw file listing.
func (pkgs *Pkgs) fetchPackage(pkgName string, sp StorePath, outname string, wg *sync.WaitGroup, count int, listings chan RawListing) {
	defer wg.Done()

	data, err := sp.FetchListing(pkgs.CacheURL, pkgs.Fetcher.StandardClient())
	if err != nil {
		log.Error("Failed to fetch listing", "name", pkgName, "err", err)
		return
	}

	listings <- RawListing{
		PkgName:    pkgName,
		OutputName: outname,
		Data:       data,
		Count:      count,
	}
}

// ProcessListings processes a channel of packages into resolved file listings.
func (pkgs *Pkgs) ProcessListings(rawPkgs chan RawListing) chan Listing {
	wg := sync.WaitGroup{}
	result := make(chan Listing)
	count := 0

	go func() {
		for raw := range rawPkgs {
			wg.Add(1)
			count++
			go pkgs.processInfo(raw, result, &wg, count)
		}

		wg.Wait()
		close(result)
	}()

	return result
}

// processInfo resolves the raw listing info to a filelist, and saves it to disk.
func (pkgs *Pkgs) processInfo(raw RawListing, listings chan Listing, wg *sync.WaitGroup, count int) {
	defer wg.Done()

	filelist := GetFileList(raw.Data)
	listings <- Listing{
		PkgName:    raw.PkgName,
		OutputName: raw.OutputName,
		Files:      filelist,
		Count:      count,
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
