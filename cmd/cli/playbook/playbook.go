package playbook

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/docker/infrakit/cmd/cli/base"
	"github.com/docker/infrakit/pkg/cli/remote"
	"github.com/docker/infrakit/pkg/discovery"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/template"
	"github.com/spf13/cobra"
)

var log = logutil.New("module", "cli/playbook")

const (
	// PlaybooksFileEnvVar is the location of the playbooks file
	PlaybooksFileEnvVar = "INFRAKIT_PLAYBOOKS_FILE"
)

func init() {
	base.Register(Command)
}

func getHome() string {
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}

func defaultPlaybooksFile() string {
	if playbooksFile := os.Getenv(PlaybooksFileEnvVar); playbooksFile != "" {
		return playbooksFile
	}
	return filepath.Join(getHome(), ".infrakit/playbooks")
}

// Load loads the playbook
func Load() (remote.Modules, error) {
	return loadPlaybooks()
}

func loadPlaybooks() (remote.Modules, error) {
	modules := remote.Modules{}
	buff, err := ioutil.ReadFile(defaultPlaybooksFile())
	if err != nil {
		if !os.IsExist(err) {
			return modules, nil
		}
		return nil, err
	}

	if len(buff) > 0 {
		err = remote.Decode(buff, &modules)
	}
	return modules, err
}

// Command is the entrypoint
func Command(plugins func() discovery.Plugins) *cobra.Command {

	///////////////////////////////////////////////////////////////////////////////////
	// playbook
	cmd := &cobra.Command{
		Use:   "playbook",
		Short: "Manage playbooks",
	}
	quiet := cmd.PersistentFlags().BoolP("quiet", "q", false, "Print rows without column headers")

	add := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a playbook",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 2 {
				cmd.Usage()
				os.Exit(1)
			}

			name := args[0]
			url := args[1]

			modules, err := loadPlaybooks()
			if err != nil {
				return err
			}

			// try fetch
			test, err := template.NewTemplate(url, template.Options{})
			if err != nil {
				return err
			}

			_, err = test.Render(nil)
			if err != nil {
				return err
			}

			if _, has := modules[remote.Op(name)]; has {
				return fmt.Errorf("%s already exists", name)
			}

			modules[remote.Op(name)] = remote.SourceURL(url)

			encoded, err := remote.Encode(modules)
			if err != nil {
				return err
			}

			return ioutil.WriteFile(defaultPlaybooksFile(), encoded, 0755)
		},
	}
	remove := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a playbook",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				cmd.Usage()
				os.Exit(1)
			}

			name := args[0]

			modules, err := loadPlaybooks()
			if err != nil {
				return err
			}

			if _, has := modules[remote.Op(name)]; !has {
				return fmt.Errorf("%s does not exists", name)
			}

			delete(modules, remote.Op(name))

			encoded, err := remote.Encode(modules)
			if err != nil {
				return err
			}

			return ioutil.WriteFile(defaultPlaybooksFile(), encoded, 0755)
		},
	}

	rawOutputFlags, rawOutput := base.RawOutput()
	list := &cobra.Command{
		Use:   "ls",
		Short: "List playbooks",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 0 {
				cmd.Usage()
				os.Exit(1)
			}

			modules, err := loadPlaybooks()
			if err != nil {
				fmt.Println("***")
				return err
			}
			rendered, err := rawOutput(os.Stdout, modules)
			if err != nil {
				return err
			}
			if rendered {
				return nil
			}

			if !*quiet {
				fmt.Printf("%-30s\t%-30s\n", "PLAYBOOK", "URL")
			}

			for op, url := range modules {
				fmt.Printf("%-30v\t%-30v\n", op, url)
			}
			return nil
		},
	}
	list.Flags().AddFlagSet(rawOutputFlags)

	cmd.AddCommand(add, remove, list)

	return cmd
}
