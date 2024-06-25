package nixpkgs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
)

// AvailableChannels returns the available channels.
func AvailableChannels(cacheDir string) ([]string, error) {
	cache := cacheDir
	if cache == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		cache = userCache
	}

	dir := cache + "/nix-hund/channels"
	d, err := os.Open(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error("Couldn't open directory", "err", err)
			return nil, err
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error("Couldn't create directory", "err", err)
			return nil, err
		}

		d, err = os.Open(dir)
		if err != nil {
			log.Error("Couldn't open directory after creation", "err", err)
			return nil, err
		}
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		log.Error("Error reading directory contents:", "err", err)
		return nil, err
	}


	result := make([]string, 0)
	for _, file := range files {
		if file.Mode().IsRegular() && strings.HasSuffix(file.Name(), ".json") {
			result = append(result, file.Name()[:len(file.Name())-5])
		}
	}

	return result, nil
}

// FetchChannel fetches the data for a specific channel. The channel should be something you can put in `nix-env --file`.
func FetchChannel(channel, outpath string) error {
	if _, err := os.Stat(outpath); err == nil {
		log.Error("File already exists", "name", outpath)
		return errors.New("already exists")
	}

	log.Info("Fetching channel data, this could take a while")

	outfile, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	buf := &bytes.Buffer{}
	cmd := exec.Command(
		"nix-env",
		"--file", channel,
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
		return err
	}

	result := list{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return err
	}

	log.Info("Fetching done", "pkgs", len(result))
	return nil
}
