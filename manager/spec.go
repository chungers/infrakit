package manager

import (
	"encoding/json"

	"github.com/docker/infrakit/plugin/group/types"
)

type pluginSpec struct {
	Plugin     string
	Properties *json.RawMessage
}

type globalSpec struct {
	Groups map[string]pluginSpec
}

func (config globalSpec) findPlugins() []string {
	// determine list of all plugins by config
	names := []string{}

	for _, plugin := range config.Groups {

		names = append(names, plugin.Plugin)

		// Try to parse the properties and if the plugin is a default group plugin then we can
		// determine the flavor and instance plugin names.
		if plugin.Properties != nil {
			spec := &types.Spec{}
			if err := json.Unmarshal([]byte(*plugin.Properties), spec); err == nil {

				if spec.Instance.Plugin != "" {
					names = append(names, spec.Instance.Plugin)
				}
				if spec.Flavor.Plugin != "" {
					names = append(names, spec.Flavor.Plugin)
				}
			}
		}
	}
	return names
}
