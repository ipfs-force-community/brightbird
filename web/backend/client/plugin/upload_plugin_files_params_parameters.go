// Code generated by go-swagger; DO NOT EDIT.

package plugin

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

// NewUploadPluginFilesParamsParams creates a new UploadPluginFilesParamsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewUploadPluginFilesParamsParams() *UploadPluginFilesParamsParams {
	return &UploadPluginFilesParamsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewUploadPluginFilesParamsParamsWithTimeout creates a new UploadPluginFilesParamsParams object
// with the ability to set a timeout on a request.
func NewUploadPluginFilesParamsParamsWithTimeout(timeout time.Duration) *UploadPluginFilesParamsParams {
	return &UploadPluginFilesParamsParams{
		timeout: timeout,
	}
}

// NewUploadPluginFilesParamsParamsWithContext creates a new UploadPluginFilesParamsParams object
// with the ability to set a context for a request.
func NewUploadPluginFilesParamsParamsWithContext(ctx context.Context) *UploadPluginFilesParamsParams {
	return &UploadPluginFilesParamsParams{
		Context: ctx,
	}
}

// NewUploadPluginFilesParamsParamsWithHTTPClient creates a new UploadPluginFilesParamsParams object
// with the ability to set a custom HTTPClient for a request.
func NewUploadPluginFilesParamsParamsWithHTTPClient(client *http.Client) *UploadPluginFilesParamsParams {
	return &UploadPluginFilesParamsParams{
		HTTPClient: client,
	}
}

/*
UploadPluginFilesParamsParams contains all the parameters to send to the API endpoint

	for the upload plugin files params operation.

	Typically these are written to a http.Request.
*/
type UploadPluginFilesParamsParams struct {

	/* Plugins.

	   Plugin file.
	*/
	PluginFiles runtime.NamedReadCloser

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the upload plugin files params params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UploadPluginFilesParamsParams) WithDefaults() *UploadPluginFilesParamsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the upload plugin files params params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UploadPluginFilesParamsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) WithTimeout(timeout time.Duration) *UploadPluginFilesParamsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) WithContext(ctx context.Context) *UploadPluginFilesParamsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) WithHTTPClient(client *http.Client) *UploadPluginFilesParamsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPluginFiles adds the plugins to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) WithPluginFiles(plugins runtime.NamedReadCloser) *UploadPluginFilesParamsParams {
	o.SetPluginFiles(plugins)
	return o
}

// SetPluginFiles adds the plugins to the upload plugin files params params
func (o *UploadPluginFilesParamsParams) SetPluginFiles(plugins runtime.NamedReadCloser) {
	o.PluginFiles = plugins
}

// WriteToRequest writes these params to a swagger request
func (o *UploadPluginFilesParamsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.PluginFiles != nil {

		if o.PluginFiles != nil {
			// form file param plugins
			if err := r.SetFileParam("plugins", o.PluginFiles); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}