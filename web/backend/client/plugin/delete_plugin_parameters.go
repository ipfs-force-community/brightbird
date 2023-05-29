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

// NewDeletePluginParams creates a new DeletePluginParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeletePluginParams() *DeletePluginParams {
	return &DeletePluginParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeletePluginParamsWithTimeout creates a new DeletePluginParams object
// with the ability to set a timeout on a request.
func NewDeletePluginParamsWithTimeout(timeout time.Duration) *DeletePluginParams {
	return &DeletePluginParams{
		timeout: timeout,
	}
}

// NewDeletePluginParamsWithContext creates a new DeletePluginParams object
// with the ability to set a context for a request.
func NewDeletePluginParamsWithContext(ctx context.Context) *DeletePluginParams {
	return &DeletePluginParams{
		Context: ctx,
	}
}

// NewDeletePluginParamsWithHTTPClient creates a new DeletePluginParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeletePluginParamsWithHTTPClient(client *http.Client) *DeletePluginParams {
	return &DeletePluginParams{
		HTTPClient: client,
	}
}

/*
DeletePluginParams contains all the parameters to send to the API endpoint

	for the delete plugin operation.

	Typically these are written to a http.Request.
*/
type DeletePluginParams struct {

	/* ID.

	   id of plugin
	*/
	ID *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete plugin params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeletePluginParams) WithDefaults() *DeletePluginParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete plugin params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeletePluginParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete plugin params
func (o *DeletePluginParams) WithTimeout(timeout time.Duration) *DeletePluginParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete plugin params
func (o *DeletePluginParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete plugin params
func (o *DeletePluginParams) WithContext(ctx context.Context) *DeletePluginParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete plugin params
func (o *DeletePluginParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete plugin params
func (o *DeletePluginParams) WithHTTPClient(client *http.Client) *DeletePluginParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete plugin params
func (o *DeletePluginParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the delete plugin params
func (o *DeletePluginParams) WithID(id *string) *DeletePluginParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the delete plugin params
func (o *DeletePluginParams) SetID(id *string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeletePluginParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.ID != nil {

		// query param id
		var qrID string

		if o.ID != nil {
			qrID = *o.ID
		}
		qID := qrID
		if qID != "" {

			if err := r.SetQueryParam("id", qID); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
