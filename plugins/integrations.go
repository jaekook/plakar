package plugins

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"runtime"
	"slices"

	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/utils"
)

type IntegrationInstallation struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

type IntegrationTypes struct {
	Storage     bool `json:"storage"`
	Source      bool `json:"source"`
	Destination bool `json:"destination"`
}

type IntegrationInfo struct {
	Id            string           `json:"id"`
	Name          string           `json:"name"`
	DisplayName   string           `json:"display_name"`
	Description   string           `json:"description"`
	Homepage      string           `json:"homepage"`
	Repository    string           `json:"repository"`
	License       string           `json:"license"`
	Tags          []string         `json:"tags"`
	APIVersion    string           `json:"api_version"`
	LatestVersion string           `json:"latest_version"`
	Types         IntegrationTypes `json:"types"`

	Documentation string `json:"documentation"` // README.md
	Icon          string `json:"icon"`          // assets/icon.{png,svg}
	Featured      string `json:"featured"`      // assets/featured.{png,svg}

	Installation IntegrationInstallation `json:"installation"`
}

type IntegrationIndex struct {
	Integrations []IntegrationInfo `json:"integrations"`
}

func fetchList() (*IntegrationIndex, error) {
	var index IntegrationIndex

	req, err := http.NewRequest("GET", "https://api.plakar.io/v1/integrations/list.json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("plakar/%s (%s/%s)", utils.VERSION, runtime.GOOS, runtime.GOARCH))
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func IterIntegrations(ctx *appcontext.AppContext, filterType string, filterTag string) iter.Seq2[*IntegrationInfo, error] {

	return func(yield func(*IntegrationInfo, error) bool) {

		index, err := fetchList()
		if err != nil {
			ctx.GetLogger().Warn("failed to retrieve integrations list %v", err)
		}

		installed, err := ListInstalledPlugins(ctx)
		if err != nil {
			ctx.GetLogger().Warn("failed to list installed plugins %v", err)
		}

		for _, info := range index.Integrations {
			if filterType == "storage" && !info.Types.Storage {
				continue
			}
			if filterType == "source" && !info.Types.Source {
				continue
			}
			if filterType == "destination" && !info.Types.Destination {
				continue
			}
			if filterTag != "" && !slices.Contains(info.Tags, filterTag) {
				continue
			}
			info.Installation.Status = "not-installed"

			plugin := installed.GetPlugin(info.Name)
			if plugin != nil {
				manifest, err := plugin.Manifest()
				if err != nil {
					ctx.GetLogger().Warn("failed to read manifest file: %v", err)
				} else {
					info.Installation.Status = "installed"
					info.Installation.Version = manifest.Version
				}
			}

			if !yield(&info, nil) {
				return
			}
		}
	}
}
