package profiles

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

// NewProfilesGetSwitchVendorParams creates a new ProfilesGetSwitchVendorParams object
// with the default values initialized.
func NewProfilesGetSwitchVendorParams() *ProfilesGetSwitchVendorParams {
	var ()
	return &ProfilesGetSwitchVendorParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewProfilesGetSwitchVendorParamsWithTimeout creates a new ProfilesGetSwitchVendorParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewProfilesGetSwitchVendorParamsWithTimeout(timeout time.Duration) *ProfilesGetSwitchVendorParams {
	var ()
	return &ProfilesGetSwitchVendorParams{

		timeout: timeout,
	}
}

// NewProfilesGetSwitchVendorParamsWithContext creates a new ProfilesGetSwitchVendorParams object
// with the default values initialized, and the ability to set a context for a request
func NewProfilesGetSwitchVendorParamsWithContext(ctx context.Context) *ProfilesGetSwitchVendorParams {
	var ()
	return &ProfilesGetSwitchVendorParams{

		Context: ctx,
	}
}

// NewProfilesGetSwitchVendorParamsWithHTTPClient creates a new ProfilesGetSwitchVendorParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewProfilesGetSwitchVendorParamsWithHTTPClient(client *http.Client) *ProfilesGetSwitchVendorParams {
	var ()
	return &ProfilesGetSwitchVendorParams{
		HTTPClient: client,
	}
}

/*ProfilesGetSwitchVendorParams contains all the parameters to send to the API endpoint
for the profiles get switch vendor operation typically these are written to a http.Request
*/
type ProfilesGetSwitchVendorParams struct {

	/*Vendor
	  The switch vendor name

	*/
	Vendor string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) WithTimeout(timeout time.Duration) *ProfilesGetSwitchVendorParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) WithContext(ctx context.Context) *ProfilesGetSwitchVendorParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) WithHTTPClient(client *http.Client) *ProfilesGetSwitchVendorParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithVendor adds the vendor to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) WithVendor(vendor string) *ProfilesGetSwitchVendorParams {
	o.SetVendor(vendor)
	return o
}

// SetVendor adds the vendor to the profiles get switch vendor params
func (o *ProfilesGetSwitchVendorParams) SetVendor(vendor string) {
	o.Vendor = vendor
}

// WriteToRequest writes these params to a swagger request
func (o *ProfilesGetSwitchVendorParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param vendor
	if err := r.SetPathParam("vendor", o.Vendor); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
