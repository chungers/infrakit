package main

import (
	"os"
	"os/user"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/cli"
	instance_plugin "github.com/docker/infrakit/rpc/instance"
	"github.com/spf13/cobra"
)

const (
	// InstanceDirEnvVar is the environment variable that may be used to customize the plugin directory
	InstanceDirEnvVar = "INFRAKIT_INSTANCE_FILE_DIR"
)

// DefaultInstanceDir is the directory where this instance plugin will store its state.
func DefaultInstanceDir() string {
	if storeDir := os.Getenv(InstanceDirEnvVar); storeDir != "" {
		return storeDir
	}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, ".infrakit/instance-file")
}

func main() {

	var logLevel int
	var name string
	var dir string

	cmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "File instance plugin",
		Run: func(c *cobra.Command, args []string) {

			cli.SetLogLevel(logLevel)
			cli.RunPlugin(name, instance_plugin.PluginServer(NewFileInstancePlugin(dir)))
		},
	}

	cmd.AddCommand(cli.VersionCommand())

	cmd.Flags().StringVar(&name, "name", filepath.Base(os.Args[0]), "Plugin name to advertise for discovery")
	cmd.PersistentFlags().IntVar(&logLevel, "log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")

	cmd.Flags().StringVar(&dir, "dir", DefaultInstanceDir(), "Dir for storing the files")

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
