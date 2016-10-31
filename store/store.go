package store

// Snapshot takes and loads snapshots of some blob.
type Snapshot interface {

	// Save saves a snapshot of the given object and revision.
	Save(obj interface{}) error

	// Load loads a snapshot and marshals into the given reference
	Load(output interface{}) error
}
