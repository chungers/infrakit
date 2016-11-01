package manager

// CommitRequest is the rpc wrapper for request parameters to Commit
type CommitRequest struct {
}

// CommitResponse is the rpc wrapper for the results of Commit
type CommitResponse struct {
	OK bool
}

// RPCService is the interface for exposing the plugin methods to rpc
type RPCService interface {

	// Commit is the rpc method for flavor Commit
	Commit(req *CommitRequest, resp *CommitResponse) error
}