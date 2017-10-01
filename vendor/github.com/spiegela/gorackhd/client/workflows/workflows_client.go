package workflows

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new workflows API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for workflows API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
WorkflowsAction performs an action on the specified workflow

Perform the specified action on the workflow with the specified instance identifier. Currently, the cancel action is supported.

*/
func (a *Client) WorkflowsAction(params *WorkflowsActionParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsActionAccepted, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsActionParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsAction",
		Method:             "PUT",
		PathPattern:        "/workflows/{identifier}/action",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsActionReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsActionAccepted), nil

}

/*
WorkflowsDeleteByInstanceID deletes the specified workflow

Delete the workflow with the specified instance identifier.
*/
func (a *Client) WorkflowsDeleteByInstanceID(params *WorkflowsDeleteByInstanceIDParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsDeleteByInstanceIDNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsDeleteByInstanceIDParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsDeleteByInstanceId",
		Method:             "DELETE",
		PathPattern:        "/workflows/{identifier}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsDeleteByInstanceIDReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsDeleteByInstanceIDNoContent), nil

}

/*
WorkflowsDeleteGraphsByName deletes the specified workflow graph

Delete the workflow graph with the specified value of the injectableName property.
*/
func (a *Client) WorkflowsDeleteGraphsByName(params *WorkflowsDeleteGraphsByNameParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsDeleteGraphsByNameNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsDeleteGraphsByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsDeleteGraphsByName",
		Method:             "DELETE",
		PathPattern:        "/workflows/graphs/{injectableName}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsDeleteGraphsByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsDeleteGraphsByNameNoContent), nil

}

/*
WorkflowsDeleteTasksByName deletes the specified workflow task

Delete the workflow task with the specified value of the injectableName property.
*/
func (a *Client) WorkflowsDeleteTasksByName(params *WorkflowsDeleteTasksByNameParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsDeleteTasksByNameNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsDeleteTasksByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsDeleteTasksByName",
		Method:             "DELETE",
		PathPattern:        "/workflows/tasks/{injectableName}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsDeleteTasksByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsDeleteTasksByNameNoContent), nil

}

/*
WorkflowsGet gets a list of workflow instances

Get list workflow that have been run or are currently running.
*/
func (a *Client) WorkflowsGet(params *WorkflowsGetParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGet",
		Method:             "GET",
		PathPattern:        "/workflows",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetOK), nil

}

/*
WorkflowsGetAllTasks gets list of workflow tasks

Get a list of all workflow tasks that can be added to a workflow.
*/
func (a *Client) WorkflowsGetAllTasks(params *WorkflowsGetAllTasksParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetAllTasksOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetAllTasksParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGetAllTasks",
		Method:             "GET",
		PathPattern:        "/workflows/tasks",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetAllTasksReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetAllTasksOK), nil

}

/*
WorkflowsGetByInstanceID gets the specified workflow

Get the workflow with the specified instance identifier.
*/
func (a *Client) WorkflowsGetByInstanceID(params *WorkflowsGetByInstanceIDParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetByInstanceIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetByInstanceIDParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGetByInstanceId",
		Method:             "GET",
		PathPattern:        "/workflows/{identifier}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetByInstanceIDReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetByInstanceIDOK), nil

}

/*
WorkflowsGetGraphs gets list of workflow graphs

Get a list of all workflow graphs available to run.
*/
func (a *Client) WorkflowsGetGraphs(params *WorkflowsGetGraphsParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetGraphsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetGraphsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGetGraphs",
		Method:             "GET",
		PathPattern:        "/workflows/graphs",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetGraphsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetGraphsOK), nil

}

/*
WorkflowsGetGraphsByName gets the specified workflow graph

Get the workflow graph with the specified value of the injectableName property.
*/
func (a *Client) WorkflowsGetGraphsByName(params *WorkflowsGetGraphsByNameParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetGraphsByNameOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetGraphsByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGetGraphsByName",
		Method:             "GET",
		PathPattern:        "/workflows/graphs/{injectableName}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetGraphsByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetGraphsByNameOK), nil

}

/*
WorkflowsGetTasksByName gets the specified workflow task

Get the task with the specified value of the injectableName property.
*/
func (a *Client) WorkflowsGetTasksByName(params *WorkflowsGetTasksByNameParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsGetTasksByNameOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsGetTasksByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsGetTasksByName",
		Method:             "GET",
		PathPattern:        "/workflows/tasks/{injectableName}",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsGetTasksByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsGetTasksByNameOK), nil

}

/*
WorkflowsPost runs a workflow

Run a workflow by specifying a workflow graph injectable name. The workflow is not associated with a node.

*/
func (a *Client) WorkflowsPost(params *WorkflowsPostParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsPostCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsPostParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsPost",
		Method:             "POST",
		PathPattern:        "/workflows",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsPostReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsPostCreated), nil

}

/*
WorkflowsPutGraphs puts a graph

Create or modify a workflow graph in the graph library.
*/
func (a *Client) WorkflowsPutGraphs(params *WorkflowsPutGraphsParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsPutGraphsCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsPutGraphsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsPutGraphs",
		Method:             "PUT",
		PathPattern:        "/workflows/graphs",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsPutGraphsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsPutGraphsCreated), nil

}

/*
WorkflowsPutTask puts a workflow task

Create or update a workflow task in the library of tasks.
*/
func (a *Client) WorkflowsPutTask(params *WorkflowsPutTaskParams, authInfo runtime.ClientAuthInfoWriter) (*WorkflowsPutTaskCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewWorkflowsPutTaskParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "workflowsPutTask",
		Method:             "PUT",
		PathPattern:        "/workflows/tasks",
		ProducesMediaTypes: []string{"application/json", "application/x-gzip"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &WorkflowsPutTaskReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*WorkflowsPutTaskCreated), nil

}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
