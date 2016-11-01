package manager

import (
	"encoding/json"

	"github.com/docker/infrakit/plugin/group/types"
	"github.com/docker/infrakit/spi/group"
)

type PluginSpec struct {
	Plugin     string
	Properties *json.RawMessage
}

type GlobalSpec struct {
	Groups map[group.ID]PluginSpec
}

func (config GlobalSpec) findPlugins() []string {
	// determine list of all plugins by config
	names := map[string]bool{}

	for _, plugin := range config.Groups {

		names[plugin.Plugin] = true

		// Try to parse the properties and if the plugin is a default group plugin then we can
		// determine the flavor and instance plugin names.
		if plugin.Properties != nil {
			spec := &types.Spec{}
			if err := json.Unmarshal([]byte(*plugin.Properties), spec); err == nil {

				if spec.Instance.Plugin != "" {
					names[spec.Instance.Plugin] = true
				}
				if spec.Flavor.Plugin != "" {
					names[spec.Flavor.Plugin] = true
				}
			}
		}
	}

	keys := []string{}
	for k := range names {
		keys = append(keys, k)
	}

	return keys
}
