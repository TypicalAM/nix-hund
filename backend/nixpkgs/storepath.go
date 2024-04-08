package nixpkgs

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	/* 	"github.com/charmbracelet/log" */
	"github.com/tidwall/gjson"
	"github.com/ulikunitz/xz"
)

// StorePath contains information about an outputs store path.
type StorePath string

// Name returns the name of the package, for example aspell-dict-qu-0.02-0.
func (sp StorePath) Name() string {
	noPrefix := string(sp[len("/nix/store/"):])
	split := strings.Split(noPrefix, "-")
	return strings.Join(split[1:], "-")
}

// Hash returns the hash of the store path for an outuput, for example qzh70f91a8sc1kb0n9hbf52hcv3jgy68.
func (sp StorePath) Hash() string {
	noPrefix := string(sp[len("/nix/store/"):])
	split := strings.Split(noPrefix, "-")
	return split[0]
}

// FileListing fetches the file listing for the package using nix binary cache provided a cache URL for example http://cache.nixos.org.
// It also does the ugly parts like decompression.
func (sp StorePath) FetchListing(url string, cli *http.Client) ([]byte, error) {
	listingURL := fmt.Sprintf("%s/%s.ls", url, sp.Hash())
	resp, err := cli.Get(listingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if slices.Contains(resp.Header["Content-Encoding"], "br") {
		// Compressed using brotli (most new pkgs)
		data, err := io.ReadAll(brotli.NewReader(resp.Body))
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	if slices.Contains(resp.Header["Content-Encoding"], "xz") {
		// Compressed using xz
		r, err := xz.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	// No compression (unlikely)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetFileList converts the binary file listing into a list of paths, for example [ /share/example ].
func GetFileList(data []byte) ([]string, error) {
	return listingHelper(string(data), "root", make([]string, 0)), nil
}

// listingHelper traverses the package listing and resolves it to paths, for example [ /share/example ].
func listingHelper(json string, cur string, result []string) []string {
	switch gjson.Get(json, cur+".type").String() {
	case "directory":
		files := make([]string, 0)

		entries := gjson.Get(json, cur+".entries")
		entries.ForEach(func(key, value gjson.Result) bool {
			sanitised := strings.ReplaceAll(key.String(), ".", `\.`)
			entry := cur + ".entries." + sanitised
			files = append(files, listingHelper(json, entry, result)...)
			return true
		})

		return files

	case "regular":
		filepath := strings.ReplaceAll(strings.ReplaceAll(cur, ".entries.", "/"), `\.`, ".")
		noPrefix := filepath[4:]
		return append(result, noPrefix)

	default:
		return make([]string, 0)
	}
}
