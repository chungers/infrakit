package swarm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/infrakit/store"
	"golang.org/x/net/context"
)

const (
	SwarmLabel = "infrakit"
)

type snapshot struct {
	client client.APIClient
}

// NewSnapshot returns an instance of the snapshot service where data is stored as a label
// in the swarm raft store.
func NewSnapshot(client client.APIClient) (store.Snapshot, error) {
	return &snapshot{client: client}, nil
}

// Save saves a snapshot of the given object and revision.
func (s *snapshot) Save(obj interface{}) error {
	label, err := encode(obj)
	if err != nil {
		return err
	}
	return writeSwarm(s.client, label)
}

// Load loads a snapshot and marshals into the given reference
func (s *snapshot) Load(output interface{}) error {
	label, err := readSwarm(s.client)
	if err != nil {
		return err
	}
	return decode(label, output)
}

func readSwarm(client client.APIClient) (string, error) {
	swarm, err := client.SwarmInspect(context.Background())
	if err != nil {
		return "", err
	}
	return swarm.ClusterInfo.Spec.Annotations.Labels[SwarmLabel], nil
}

func writeSwarm(client client.APIClient, value string) error {
	info, err := client.SwarmInspect(context.Background())
	if err != nil {
		return err
	}
	info.ClusterInfo.Spec.Annotations.Labels[SwarmLabel] = value
	return client.SwarmUpdate(context.Background(), info.ClusterInfo.Meta.Version, info.ClusterInfo.Spec,
		swarm.UpdateFlags{})
}

func encode(obj interface{}) (string, error) {
	var label bytes.Buffer
	// encoding chain
	b64 := base64.NewEncoder(base64.StdEncoding, &label)
	jsonw := json.NewEncoder(b64)
	jsonw.SetIndent("", "  ")
	err := jsonw.Encode(obj)
	if err != nil {
		return "", err
	}
	return label.String(), nil
}

func decode(label string, output interface{}) error {
	// decoding chain
	input := bytes.NewBufferString(label)
	b64 := base64.NewDecoder(base64.StdEncoding, input)
	jsonr := json.NewDecoder(b64)
	return jsonr.Decode(output)
}
