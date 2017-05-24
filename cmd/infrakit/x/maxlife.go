package x

import (
	"os"
	"strings"
	"time"

	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/plugin"
	instance_rpc "github.com/docker/infrakit/pkg/rpc/instance"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/x/maxlife"
	"github.com/spf13/cobra"
)

func maxlifeCommand(plugins func() discovery.Plugins) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "maxlife <instance plugin name>...",
		Short: "Sets max life on the given instances",
	}

	//name := cmd.Flags().String("name", "", "Name to use as name of this plugin")
	poll := cmd.Flags().DurationP("poll", "i", 10*time.Second, "Polling interval")
	maxlifeDuration := cmd.Flags().DurationP("maxlife", "m", 10*time.Minute, "Max lifetime of the resource")
	flagTags := cmd.Flags().StringSliceP("tag", "t", []string{}, "Tags to filter instance by")

	cmd.RunE = func(c *cobra.Command, args []string) error {

		if len(args) == 0 {
			cmd.Usage()
			os.Exit(-1)
		}

		tags := toTags(*flagTags)

		// Now we have a list of instance plugins to maxlife
		plugins, err := getInstancePlugins(plugins, args)
		if err != nil {
			return err
		}

		// For each we start a goroutine to poll and kill instances
		controllers := []*maxlife.Controller{}

		for name, plugin := range plugins {

			controller := maxlife.NewController(name, plugin, *poll, *maxlifeDuration, tags)
			controller.Start()

			controllers = append(controllers, controller)
		}

		// TODO - publish events when we start taking down instances.
		done := make(chan struct{})

		<-done
		return nil
	}

	return cmd
}

func ensureMaxlife(name string, plugin instance.Plugin, stop chan struct{}, poll, maxlife time.Duration,
	tags map[string]string, initialCount int) {
}
func getInstancePlugins(plugins func() discovery.Plugins, names []string) (map[string]instance.Plugin, error) {
	targets := map[string]instance.Plugin{}
	for _, target := range names {
		endpoint, err := plugins().Find(plugin.Name(target))
		if err != nil {
			return nil, err
		}
		if p, err := instance_rpc.NewClient(plugin.Name(target), endpoint.Address); err == nil {
			targets[target] = p
		} else {
			return nil, err
		}
	}
	return targets, nil
}

func toTags(slice []string) map[string]string {
	tags := map[string]string{}

	for _, tag := range slice {
		kv := strings.SplitN(tag, "=", 2)
		if len(kv) != 2 {
			log.Warn("bad format tag", "input", tag)
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if key != "" && val != "" {
			tags[key] = val
		}
	}
	return tags
}
