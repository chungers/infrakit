package nodes

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewNodesGetObmsByNodeIDParams creates a new NodesGetObmsByNodeIDParams object
// with the default values initialized.
func NewNodesGetObmsByNodeIDParams() *NodesGetObmsByNodeIDParams {
	var ()
	return &NodesGetObmsByNodeIDParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewNodesGetObmsByNodeIDParamsWithTimeout creates a new NodesGetObmsByNodeIDParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewNodesGetObmsByNodeIDParamsWithTimeout(timeout time.Duration) *NodesGetObmsByNodeIDParams {
	var ()
	return &NodesGetObmsByNodeIDParams{

		timeout: timeout,
	}
}

// NewNodesGetObmsByNodeIDParamsWithContext creates a new NodesGetObmsByNodeIDParams object
// with the default values initialized, and the ability to set a context for a request
func NewNodesGetObmsByNodeIDParamsWithContext(ctx context.Context) *NodesGetObmsByNodeIDParams {
	var ()
	return &NodesGetObmsByNodeIDParams{

		Context: ctx,
	}
}

// NewNodesGetObmsByNodeIDParamsWithHTTPClient creates a new NodesGetObmsByNodeIDParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewNodesGetObmsByNodeIDParamsWithHTTPClient(client *http.Client) *NodesGetObmsByNodeIDParams {
	var ()
	return &NodesGetObmsByNodeIDParams{
		HTTPClient: client,
	}
}

/*NodesGetObmsByNodeIDParams contains all the parameters to send to the API endpoint
for the nodes get obms by node Id operation typically these are written to a http.Request
*/
type NodesGetObmsByNodeIDParams struct {

	/*Identifier
	  The Node identifier

	*/
	Identifier string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) WithTimeout(timeout time.Duration) *NodesGetObmsByNodeIDParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) WithContext(ctx context.Context) *NodesGetObmsByNodeIDParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) WithHTTPClient(client *http.Client) *NodesGetObmsByNodeIDParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithIdentifier adds the identifier to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) WithIdentifier(identifier string) *NodesGetObmsByNodeIDParams {
	o.SetIdentifier(identifier)
	return o
}

// SetIdentifier adds the identifier to the nodes get obms by node Id params
func (o *NodesGetObmsByNodeIDParams) SetIdentifier(identifier string) {
	o.Identifier = identifier
}

// WriteToRequest writes these params to a swagger request
func (o *NodesGetObmsByNodeIDParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param identifier
	if err := r.SetPathParam("identifier", o.Identifier); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
