package plugins

import (
	_ "embed"
	"encoding/json"
	"iter"
	"slices"

	"github.com/PlakarKorp/plakar/appcontext"
)

type IntegrationTypes struct {
	Storage     bool `json:"storage"`
	Source      bool `json:"source"`
	Destination bool `json:"destination"`
}

type IntegrationStatus struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
}

type IntegrationInfo struct {
	Id            string            `json:"id"`
	DispayName    string            `json:"dispay_name"`
	Description   string            `json:"description"`
	Documentation string            `json:"documentation"`
	Icon          string            `json:"icon"`
	Logo          string            `json:"logo"`
	Tags          []string          `json:"tags"`
	LatestVersion string            `json:"latest_version"`
	Types         IntegrationTypes  `json:"types"`
	Status        IntegrationStatus `json:"status"`
}

//go:embed integrations.json
var integrationsData []byte

func IterIntegrations(ctx *appcontext.AppContext, filterType string, filterTag string) iter.Seq2[*IntegrationInfo, error] {
	return func(yield func(*IntegrationInfo, error) bool) {
		var list []IntegrationInfo
		err := json.Unmarshal(integrationsData, &list)
		if err != nil {
			yield(nil, err)
			return
		}

		for _, info := range list {
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
			// XXX Check if installed
			if !yield(&info, nil) {
				return
			}
		}
	}
}
