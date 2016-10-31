package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/docker/infrakit/store"
)

const (
	// StoreDirEnvVar is the directory where the configs are stored
	StoreDirEnvVar = "INFRAKIT_STORE_DIR"
)

// DefaultLaderFile is the file that this detector uses to decide who the leader is.
// In a mult-host set up, it's assumed that the file system would be share (e.g. NFS mount or S3 FUSE etc.)
func DefaultStoreDir() string {
	if storeDir := os.Getenv(StoreDirEnvVar); storeDir != "" {
		return storeDir
	}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, ".infrakit/config")
}

type snapshot struct {
	dir  string
	name string
}

// NewSnapshot returns an instance of the snapshot service where data is stored in the directory given.
// This is a simple implementation that assumes a single file for the entire snapshot.
func NewSnapshot(dir, name string) (store.Snapshot, error) {
	// file must exist
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("file %s must be a directory", dir)
	}

	return &snapshot{dir: dir, name: name}, nil
}

// Save saves a snapshot of the given object and revision.
func (s *snapshot) Save(obj interface{}) error {
	buff, err := json.MarshalIndent(obj, "  ", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(s.dir, s.name), buff, 0644)
}

// Load loads a snapshot and marshals into the given reference
func (s *snapshot) Load(output interface{}) error {
	buff, err := ioutil.ReadFile(filepath.Join(s.dir, s.name))
	if err != nil {
		return err
	}
	return json.Unmarshal(buff, output)
}
