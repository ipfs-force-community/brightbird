// Code generated by go-swagger; DO NOT EDIT.

package task

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

// NewGetTaskReqParams creates a new GetTaskReqParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetTaskReqParams() *GetTaskReqParams {
	return &GetTaskReqParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetTaskReqParamsWithTimeout creates a new GetTaskReqParams object
// with the ability to set a timeout on a request.
func NewGetTaskReqParamsWithTimeout(timeout time.Duration) *GetTaskReqParams {
	return &GetTaskReqParams{
		timeout: timeout,
	}
}

// NewGetTaskReqParamsWithContext creates a new GetTaskReqParams object
// with the ability to set a context for a request.
func NewGetTaskReqParamsWithContext(ctx context.Context) *GetTaskReqParams {
	return &GetTaskReqParams{
		Context: ctx,
	}
}

// NewGetTaskReqParamsWithHTTPClient creates a new GetTaskReqParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetTaskReqParamsWithHTTPClient(client *http.Client) *GetTaskReqParams {
	return &GetTaskReqParams{
		HTTPClient: client,
	}
}

/*
GetTaskReqParams contains all the parameters to send to the API endpoint

	for the get task req operation.

	Typically these are written to a http.Request.
*/
type GetTaskReqParams struct {

	// ID.
	ID *string

	// TestID.
	TestID *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get task req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskReqParams) WithDefaults() *GetTaskReqParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get task req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskReqParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get task req params
func (o *GetTaskReqParams) WithTimeout(timeout time.Duration) *GetTaskReqParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get task req params
func (o *GetTaskReqParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get task req params
func (o *GetTaskReqParams) WithContext(ctx context.Context) *GetTaskReqParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get task req params
func (o *GetTaskReqParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get task req params
func (o *GetTaskReqParams) WithHTTPClient(client *http.Client) *GetTaskReqParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get task req params
func (o *GetTaskReqParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the get task req params
func (o *GetTaskReqParams) WithID(id *string) *GetTaskReqParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get task req params
func (o *GetTaskReqParams) SetID(id *string) {
	o.ID = id
}

// WithTestID adds the testID to the get task req params
func (o *GetTaskReqParams) WithTestID(testID *string) *GetTaskReqParams {
	o.SetTestID(testID)
	return o
}

// SetTestID adds the testId to the get task req params
func (o *GetTaskReqParams) SetTestID(testID *string) {
	o.TestID = testID
}

// WriteToRequest writes these params to a swagger request
func (o *GetTaskReqParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.ID != nil {

		// query param ID
		var qrID string

		if o.ID != nil {
			qrID = *o.ID
		}
		qID := qrID
		if qID != "" {

			if err := r.SetQueryParam("ID", qID); err != nil {
				return err
			}
		}
	}

	if o.TestID != nil {

		// query param testID
		var qrTestID string

		if o.TestID != nil {
			qrTestID = *o.TestID
		}
		qTestID := qrTestID
		if qTestID != "" {

			if err := r.SetQueryParam("testID", qTestID); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
