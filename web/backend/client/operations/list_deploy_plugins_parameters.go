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

// NewListDeployPluginsParams creates a new ListDeployPluginsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListDeployPluginsParams() *ListDeployPluginsParams {
	return &ListDeployPluginsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListDeployPluginsParamsWithTimeout creates a new ListDeployPluginsParams object
// with the ability to set a timeout on a request.
func NewListDeployPluginsParamsWithTimeout(timeout time.Duration) *ListDeployPluginsParams {
	return &ListDeployPluginsParams{
		timeout: timeout,
	}
}

// NewListDeployPluginsParamsWithContext creates a new ListDeployPluginsParams object
// with the ability to set a context for a request.
func NewListDeployPluginsParamsWithContext(ctx context.Context) *ListDeployPluginsParams {
	return &ListDeployPluginsParams{
		Context: ctx,
	}
}

// NewListDeployPluginsParamsWithHTTPClient creates a new ListDeployPluginsParams object
// with the ability to set a custom HTTPClient for a request.
func NewListDeployPluginsParamsWithHTTPClient(client *http.Client) *ListDeployPluginsParams {
	return &ListDeployPluginsParams{
		HTTPClient: client,
	}
}

/*
ListDeployPluginsParams contains all the parameters to send to the API endpoint

	for the list deploy plugins operation.

	Typically these are written to a http.Request.
*/
type ListDeployPluginsParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list deploy plugins params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListDeployPluginsParams) WithDefaults() *ListDeployPluginsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list deploy plugins params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListDeployPluginsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list deploy plugins params
func (o *ListDeployPluginsParams) WithTimeout(timeout time.Duration) *ListDeployPluginsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list deploy plugins params
func (o *ListDeployPluginsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list deploy plugins params
func (o *ListDeployPluginsParams) WithContext(ctx context.Context) *ListDeployPluginsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list deploy plugins params
func (o *ListDeployPluginsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list deploy plugins params
func (o *ListDeployPluginsParams) WithHTTPClient(client *http.Client) *ListDeployPluginsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list deploy plugins params
func (o *ListDeployPluginsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListDeployPluginsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
