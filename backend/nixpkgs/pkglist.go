package nixpkgs

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/TypicalAM/nix-hund/metrics"
	"github.com/charmbracelet/log"
	retryhttp "github.com/hashicorp/go-retryablehttp"
)

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

// New reads or fetches the available packages from nixpkgs. It uses the specified channel and the cache url provided by the caller. Use `nixpkgs.AvailableChannels()` to get available channels.
func New(url, channel string) (*Pkgs, error) {
	cli := retryhttp.NewClient()
	cli.RetryMax = 5
	cli.Backoff = retryhttp.LinearJitterBackoff
	cli.Logger = log.New(io.Discard)

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	dir := cache + "/nix-hund/channels/" + channel + ".json"
	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}

	pkgs := make(list)
	if err := json.Unmarshal(data, &pkgs); err != nil {
		return nil, err
	}

	metrics.PackageCount.Set(float64(len(pkgs)))
	return &Pkgs{
		CacheURL: url,
		List:     pkgs,
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
	OutputHash string
	Version    string
	Data       []byte
	Count      int
}

// Listing is a listing broken down into individual files.
type Listing struct {
	PkgName    string
	OutputName string
	OutputHash string
	Version    string
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
			go pkgs.fetchPackage(pkgName, sp, pkg.Version, outname, &wg, count, rawListings)
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
func (pkgs *Pkgs) fetchPackage(pkgName string, sp StorePath, version string, outname string, wg *sync.WaitGroup, count int, listings chan RawListing) {
	defer wg.Done()

	data, err := sp.FetchListing(pkgs.CacheURL, pkgs.Fetcher.StandardClient())
	if err != nil {
		log.Error("Failed to fetch listing", "name", pkgName, "err", err)
		return
	}

	listings <- RawListing{
		PkgName:    pkgName,
		OutputName: outname,
		OutputHash: sp.Hash(),
		Version:    version,
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
		OutputHash: raw.OutputHash,
		Version:    raw.Version,
		Files:      filelist,
		Count:      count,
	}
}
