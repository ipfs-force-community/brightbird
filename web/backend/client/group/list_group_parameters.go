// Code generated by go-swagger; DO NOT EDIT.

package group

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

// NewListGroupParams creates a new ListGroupParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListGroupParams() *ListGroupParams {
	return &ListGroupParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListGroupParamsWithTimeout creates a new ListGroupParams object
// with the ability to set a timeout on a request.
func NewListGroupParamsWithTimeout(timeout time.Duration) *ListGroupParams {
	return &ListGroupParams{
		timeout: timeout,
	}
}

// NewListGroupParamsWithContext creates a new ListGroupParams object
// with the ability to set a context for a request.
func NewListGroupParamsWithContext(ctx context.Context) *ListGroupParams {
	return &ListGroupParams{
		Context: ctx,
	}
}

// NewListGroupParamsWithHTTPClient creates a new ListGroupParams object
// with the ability to set a custom HTTPClient for a request.
func NewListGroupParamsWithHTTPClient(client *http.Client) *ListGroupParams {
	return &ListGroupParams{
		HTTPClient: client,
	}
}

/*
ListGroupParams contains all the parameters to send to the API endpoint

	for the list group operation.

	Typically these are written to a http.Request.
*/
type ListGroupParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListGroupParams) WithDefaults() *ListGroupParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListGroupParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list group params
func (o *ListGroupParams) WithTimeout(timeout time.Duration) *ListGroupParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list group params
func (o *ListGroupParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list group params
func (o *ListGroupParams) WithContext(ctx context.Context) *ListGroupParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list group params
func (o *ListGroupParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list group params
func (o *ListGroupParams) WithHTTPClient(client *http.Client) *ListGroupParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list group params
func (o *ListGroupParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListGroupParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
