package templates

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

// NewTemplatesMetaGetByNameParams creates a new TemplatesMetaGetByNameParams object
// with the default values initialized.
func NewTemplatesMetaGetByNameParams() *TemplatesMetaGetByNameParams {
	var ()
	return &TemplatesMetaGetByNameParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewTemplatesMetaGetByNameParamsWithTimeout creates a new TemplatesMetaGetByNameParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewTemplatesMetaGetByNameParamsWithTimeout(timeout time.Duration) *TemplatesMetaGetByNameParams {
	var ()
	return &TemplatesMetaGetByNameParams{

		timeout: timeout,
	}
}

// NewTemplatesMetaGetByNameParamsWithContext creates a new TemplatesMetaGetByNameParams object
// with the default values initialized, and the ability to set a context for a request
func NewTemplatesMetaGetByNameParamsWithContext(ctx context.Context) *TemplatesMetaGetByNameParams {
	var ()
	return &TemplatesMetaGetByNameParams{

		Context: ctx,
	}
}

// NewTemplatesMetaGetByNameParamsWithHTTPClient creates a new TemplatesMetaGetByNameParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewTemplatesMetaGetByNameParamsWithHTTPClient(client *http.Client) *TemplatesMetaGetByNameParams {
	var ()
	return &TemplatesMetaGetByNameParams{
		HTTPClient: client,
	}
}

/*TemplatesMetaGetByNameParams contains all the parameters to send to the API endpoint
for the templates meta get by name operation typically these are written to a http.Request
*/
type TemplatesMetaGetByNameParams struct {

	/*Name
	  The file name of the template

	*/
	Name string
	/*Scope
	  The template scope

	*/
	Scope *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) WithTimeout(timeout time.Duration) *TemplatesMetaGetByNameParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) WithContext(ctx context.Context) *TemplatesMetaGetByNameParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) WithHTTPClient(client *http.Client) *TemplatesMetaGetByNameParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) WithName(name string) *TemplatesMetaGetByNameParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) SetName(name string) {
	o.Name = name
}

// WithScope adds the scope to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) WithScope(scope *string) *TemplatesMetaGetByNameParams {
	o.SetScope(scope)
	return o
}

// SetScope adds the scope to the templates meta get by name params
func (o *TemplatesMetaGetByNameParams) SetScope(scope *string) {
	o.Scope = scope
}

// WriteToRequest writes these params to a swagger request
func (o *TemplatesMetaGetByNameParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if o.Scope != nil {

		// query param scope
		var qrScope string
		if o.Scope != nil {
			qrScope = *o.Scope
		}
		qScope := qrScope
		if qScope != "" {
			if err := r.SetQueryParam("scope", qScope); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
