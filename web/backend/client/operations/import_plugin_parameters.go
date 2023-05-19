// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewImportPluginParams creates a new ImportPluginParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewImportPluginParams() *ImportPluginParams {
	return &ImportPluginParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewImportPluginParamsWithTimeout creates a new ImportPluginParams object
// with the ability to set a timeout on a request.
func NewImportPluginParamsWithTimeout(timeout time.Duration) *ImportPluginParams {
	return &ImportPluginParams{
		timeout: timeout,
	}
}

// NewImportPluginParamsWithContext creates a new ImportPluginParams object
// with the ability to set a context for a request.
func NewImportPluginParamsWithContext(ctx context.Context) *ImportPluginParams {
	return &ImportPluginParams{
		Context: ctx,
	}
}

// NewImportPluginParamsWithHTTPClient creates a new ImportPluginParams object
// with the ability to set a custom HTTPClient for a request.
func NewImportPluginParamsWithHTTPClient(client *http.Client) *ImportPluginParams {
	return &ImportPluginParams{
		HTTPClient: client,
	}
}

/*
ImportPluginParams contains all the parameters to send to the API endpoint

	for the import plugin operation.

	Typically these are written to a http.Request.
*/
type ImportPluginParams struct {

	/* Path.

	   directory of plugins
	*/
	Path string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the import plugin params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ImportPluginParams) WithDefaults() *ImportPluginParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the import plugin params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ImportPluginParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the import plugin params
func (o *ImportPluginParams) WithTimeout(timeout time.Duration) *ImportPluginParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the import plugin params
func (o *ImportPluginParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the import plugin params
func (o *ImportPluginParams) WithContext(ctx context.Context) *ImportPluginParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the import plugin params
func (o *ImportPluginParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the import plugin params
func (o *ImportPluginParams) WithHTTPClient(client *http.Client) *ImportPluginParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the import plugin params
func (o *ImportPluginParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPath adds the path to the import plugin params
func (o *ImportPluginParams) WithPath(path string) *ImportPluginParams {
	o.SetPath(path)
	return o
}

// SetPath adds the path to the import plugin params
func (o *ImportPluginParams) SetPath(path string) {
	o.Path = path
}

// WriteToRequest writes these params to a swagger request
func (o *ImportPluginParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param path
	qrPath := o.Path
	qPath := qrPath
	if qPath != "" {

		if err := r.SetQueryParam("path", qPath); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
