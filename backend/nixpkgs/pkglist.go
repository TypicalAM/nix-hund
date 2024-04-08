package nixpkgs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// Info is the information about an available nixpkgs derivation.
type Info struct {
	Name       string               `json:"name"`
	OutputName string               `json:"outputName"`
	Outputs    map[string]StorePath `json:"outputs"`
	Pname      string               `json:"pname"`
	System     string               `json:"system"`
	Version    string               `json:"version"`
}

// List is the nixpkgs package list.
type List map[string]Info

// Pkgs is the package list fetcher
type Pkgs struct {
	CacheURL string
	List     List
	Fetcher  *http.Client
}

// New reads or fetches the available packages from nixpkgs. It uses the default system channel and the cache url provided by the caller.
func New(url string) (*Pkgs, error) {
	cli := retryhttp.NewClient()
	cli.RetryMax = 5
	cli.Backoff = retryhttp.LinearJitterBackoff
	cli.Logger = log.New(io.Discard) // Why is it shouting so much, shut up!

	list, err := listFromCache()
	if err == nil {
		return &Pkgs{
			CacheURL: url,
			List:     list,
			Fetcher:  cli.StandardClient(),
		}, nil
	}

	list, err = fetchList()
	if err != nil {
		return nil, err
	}

	return &Pkgs{
		CacheURL: url,
		List:     list,
		Fetcher:  cli.StandardClient(),
	}, nil
}

// Count returns the total number of all derivations (NOT all outputs).
func (pkgs Pkgs) Count() int {
	return len(pkgs.List)
}

// rawListing is the information about a listing.
type rawListing struct {
	pkgName    string
	outputName string
	data       []byte
	count      int
}

// processedListing is a listing broken down into individual files.
type processedListing struct {
	pkgName    string
	outputName string
	files      []string
	count      int
}

// CreateIndex takes all the queried packages and fetches file listings for them.
func (pkgs *Pkgs) CreateIndex() error {
	log.Info("Creating index")

	totalCount := 0
	for _, pkg := range pkgs.List {
		for outname := range pkg.Outputs {
			if outname == "dev" {
				totalCount++
			}
		}
	}

	fetchCount := 0
	fetchWg := sync.WaitGroup{}
	fetched := make(chan rawListing)

	for _, pkg := range pkgs.List {
		for outname, sp := range pkg.Outputs {
			if outname != "dev" {
				continue
			}

			fetchWg.Add(1)
			fetchCount++
			log.Info("Fetching package", "name", sp.Name(), "outname", outname)
			go pkgs.fetchPackage(&fetchWg, fetchCount, outname, sp, fetched)
		}
	}

	go func() {
		fetchWg.Wait()
		close(fetched)
		log.Info("Fetching done")
	}()

	log.Info("Fetching listings queued", "queued", fetchCount)

	processedCount := 0
	processedWg := sync.WaitGroup{}
	processed := make(chan processedListing)

	for raw := range fetched {
		processedWg.Add(1)
		processedCount++
		go pkgs.processInfo(&processedWg, processedCount, raw, processed)
	}

	go func() {
		processedWg.Wait()
		close(processed)
		log.Info("Processing done")
	}()

	log.Info("Started processing")
	for info := range processed {
		log.Info("Package info", "name", info.pkgName, "outname", info.outputName, "file_count", len(info.files))
	}

	log.Info("Done processing listings", "pkgs", processedCount)
	return nil
}

// fetchPackage fetches a raw file listing.
func (pkgs *Pkgs) fetchPackage(wg *sync.WaitGroup, count int, outputName string, sp StorePath, listings chan rawListing) {
	defer wg.Done()

	data, err := sp.FetchListing(pkgs.CacheURL, pkgs.Fetcher)
	if err != nil {
		log.Error("Failed to fetch listing", "name", sp.Name(), "err", err)
		return
	}

	listings <- rawListing{
		outputName: outputName,
		pkgName:    sp.Name(),
		data:       data,
		count:      count,
	}
}

// processInfo resolves the raw listing info to a filelist, and saves it to disk.
func (pkgs *Pkgs) processInfo(wg *sync.WaitGroup, count int, raw rawListing, listings chan processedListing) {
	defer wg.Done()

	files, err := GetFileList(raw.data)
	if err != nil {
		log.Error("Failed to fetch listing", "name", raw.pkgName, "err", err)
		return
	}

	listings <- processedListing{
		pkgName:    raw.pkgName,
		outputName: raw.outputName,
		files:      files,
		count:      count,
	}
}

// listFromCache gets the package list from the user's cache directory.
func listFromCache() (List, error) {
	log.Debug("Trying to find cache")
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := dir + "/" + APP_NAME
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err2 := os.Mkdir(path, 0666); err2 != nil {
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

	result := List{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// fetchList fetches the package list from the reemote server.
func fetchList() (List, error) {
	log.Info("Fetching pkg lists")

	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := dir + "/" + APP_NAME
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

	filename := path + "/" + time.Now().Format(time.DateTime)
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

	result := List{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}

	log.Info("Fetching done", "pkgs", len(result))
	return result, nil
}
