// Code generated by go-swagger; DO NOT EDIT.

package failed_tasks

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

// NewListFailedTasksReqParams creates a new ListFailedTasksReqParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListFailedTasksReqParams() *ListFailedTasksReqParams {
	return &ListFailedTasksReqParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListFailedTasksReqParamsWithTimeout creates a new ListFailedTasksReqParams object
// with the ability to set a timeout on a request.
func NewListFailedTasksReqParamsWithTimeout(timeout time.Duration) *ListFailedTasksReqParams {
	return &ListFailedTasksReqParams{
		timeout: timeout,
	}
}

// NewListFailedTasksReqParamsWithContext creates a new ListFailedTasksReqParams object
// with the ability to set a context for a request.
func NewListFailedTasksReqParamsWithContext(ctx context.Context) *ListFailedTasksReqParams {
	return &ListFailedTasksReqParams{
		Context: ctx,
	}
}

// NewListFailedTasksReqParamsWithHTTPClient creates a new ListFailedTasksReqParams object
// with the ability to set a custom HTTPClient for a request.
func NewListFailedTasksReqParamsWithHTTPClient(client *http.Client) *ListFailedTasksReqParams {
	return &ListFailedTasksReqParams{
		HTTPClient: client,
	}
}

/*
ListFailedTasksReqParams contains all the parameters to send to the API endpoint

	for the list failed tasks req operation.

	Typically these are written to a http.Request.
*/
type ListFailedTasksReqParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list failed tasks req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListFailedTasksReqParams) WithDefaults() *ListFailedTasksReqParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list failed tasks req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListFailedTasksReqParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list failed tasks req params
func (o *ListFailedTasksReqParams) WithTimeout(timeout time.Duration) *ListFailedTasksReqParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list failed tasks req params
func (o *ListFailedTasksReqParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list failed tasks req params
func (o *ListFailedTasksReqParams) WithContext(ctx context.Context) *ListFailedTasksReqParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list failed tasks req params
func (o *ListFailedTasksReqParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list failed tasks req params
func (o *ListFailedTasksReqParams) WithHTTPClient(client *http.Client) *ListFailedTasksReqParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list failed tasks req params
func (o *ListFailedTasksReqParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListFailedTasksReqParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
