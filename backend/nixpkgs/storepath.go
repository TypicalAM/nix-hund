package nixpkgs

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
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
func GetFileList(data []byte) []string {
	result := make([]string, 0)
	root, _ := sonic.Get(data)

	queue := [][]interface{}{{"root"}}
	for len(queue) != 0 {
		path := queue[0]
		queue = queue[1:]

		item := root.GetByPath(path...)
		filetype, _ := item.Get("type").String()
		switch filetype {
		case "directory":
			items, _ := item.Get("entries").Map()
			for key := range items {
				newPath := make([]interface{}, len(path), len(path)+2)
				copy(newPath, path)
				newPath = append(newPath, "entries", key)
				queue = append(queue, newPath)
			}

			continue

		case "regular":
			filepath := ""
			for _, elem := range path {
				if elem != "entries" {
					filepath += "/" + elem.(string)
				}
			}

			result = append(result, filepath[5:]) // NOTE: We are cutting /root from the filepath
			continue
		}
	}

	return result
}
