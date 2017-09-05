package controller

import (
	"fmt"
	"io"
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
	"github.com/spf13/cobra"
)

// Specs returns the specs command
func Specs(name string, services *cli.Services) *cobra.Command {
	specs := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect returns the desired state / Specs for all managed objects",
	}
	specs.Flags().AddFlagSet(services.OutputFlags)

	tags := specs.Flags().StringSlice("tags", []string{}, "Tags to filter")

	specs.RunE = func(cmd *cobra.Command, args []string) error {

		pluginName := plugin.Name(name)
		_, objectName := pluginName.GetLookupAndType()
		if objectName == "" {
			if len(args) < 1 {
				objectName = ""
				// cmd.Usage()
				// os.Exit(1)

			} else {
				objectName = args[0]
			}
		}

		controller, err := Load(services.Plugins(), name)
		if err != nil {
			return nil
		}
		cli.MustNotNil(controller, "controller not found", "name", name)

		search := (types.Metadata{
			Name: objectName,
		}).AddTagsFromStringSlice(*tags)

		q := &search
		if q.Name == "" {
			q = nil // select all if nil
		}

		specs, err := controller.Specs(q)
		if err != nil {
			return err
		}

		return services.Output(os.Stdout, specs,
			func(w io.Writer, v interface{}) error {
				for _, s := range specs {
					fmt.Printf("%-10s  %-15s\n", "KIND", "NAME")
					fmt.Printf("%-10s  %-15s\n", s.Kind, s.Metadata.Name)
				}
				return nil
			})
	}
	return specs
}
