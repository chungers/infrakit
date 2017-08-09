package manager

import (
	"path/filepath"
	"time"

	file_leader "github.com/docker/infrakit/pkg/leader/file"
	"github.com/docker/infrakit/pkg/run"
	file_store "github.com/docker/infrakit/pkg/store/file"
	"github.com/docker/infrakit/pkg/types"
)

const (
	// EnvLeaderFile is the environment variable that may be used to customize the plugin leader detection
	EnvLeaderFile = "INFRAKIT_LEADER_FILE"

	// EnvStoreDir is the directory where the configs are stored
	EnvStoreDir = "INFRAKIT_STORE_DIR"

	// EnvURL is the location of this node
	EnvURL = "INFRAKIT_URL"
)

// BackendFileOptions contain the options for the file backend
type BackendFileOptions struct {
	// PollInterval is how often to check
	PollInterval time.Duration

	// LeaderFile is the location of the leader file
	LeaderFile string

	// StoreDir is the path to the directory where state is stored
	StoreDir string

	// ID is the id of the node
	ID string

	// LeaderLocationFile is the path to the file that stores the leader's location
	LeaderLocationFile string
}

// DefaultBackendFileOptions is the default for the file backend
var DefaultBackendFileOptions = Options{
	Backend: "file",
	Settings: types.AnyValueMust(
		BackendFileOptions{
			ID:                 "manager1",
			PollInterval:       5 * time.Second,
			LeaderFile:         run.GetEnv(EnvLeaderFile, filepath.Join(run.InfrakitHome(), "leader")),
			StoreDir:           run.GetEnv(EnvStoreDir, filepath.Join(run.InfrakitHome(), "configs")),
			LeaderLocationFile: run.GetEnv(EnvStoreDir, filepath.Join(run.InfrakitHome(), "leader.loc")),
		},
	),
	Mux: &MuxConfig{
		Listen:       ":24864",
		URL:          run.GetEnv(EnvURL, "http://localhost:24864"),
		PollInterval: 5 * time.Second,
	},
}

func configFileBackends(options BackendFileOptions, managerConfig *Options, muxConfig *MuxConfig) error {

	leader, err := file_leader.NewDetector(options.PollInterval, options.LeaderFile, options.ID)
	if err != nil {
		return err
	}

	snapshot, err := file_store.NewSnapshot(options.StoreDir, "global.config")
	if err != nil {
		return err
	}

	if managerConfig != nil {
		managerConfig.leader = leader
		managerConfig.store = snapshot
	}

	if muxConfig != nil {

		poller, err := file_leader.NewDetector(muxConfig.PollInterval, options.LeaderFile, options.ID)
		if err != nil {
			return err
		}

		muxConfig.poller = poller
		muxConfig.store = file_leader.NewStore(options.LeaderLocationFile)
	}

	return nil
}
