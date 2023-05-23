// Code generated by go-swagger; DO NOT EDIT.

package testflow

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

	"github.com/hunjixin/brightbird/models"
)

// NewSaveTestFlowParams creates a new SaveTestFlowParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewSaveTestFlowParams() *SaveTestFlowParams {
	return &SaveTestFlowParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewSaveTestFlowParamsWithTimeout creates a new SaveTestFlowParams object
// with the ability to set a timeout on a request.
func NewSaveTestFlowParamsWithTimeout(timeout time.Duration) *SaveTestFlowParams {
	return &SaveTestFlowParams{
		timeout: timeout,
	}
}

// NewSaveTestFlowParamsWithContext creates a new SaveTestFlowParams object
// with the ability to set a context for a request.
func NewSaveTestFlowParamsWithContext(ctx context.Context) *SaveTestFlowParams {
	return &SaveTestFlowParams{
		Context: ctx,
	}
}

// NewSaveTestFlowParamsWithHTTPClient creates a new SaveTestFlowParams object
// with the ability to set a custom HTTPClient for a request.
func NewSaveTestFlowParamsWithHTTPClient(client *http.Client) *SaveTestFlowParams {
	return &SaveTestFlowParams{
		HTTPClient: client,
	}
}

/*
SaveTestFlowParams contains all the parameters to send to the API endpoint

	for the save test flow operation.

	Typically these are written to a http.Request.
*/
type SaveTestFlowParams struct {

	/* Testflow.

	   test flow json
	*/
	Testflow *models.TestFlow

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the save test flow params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SaveTestFlowParams) WithDefaults() *SaveTestFlowParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the save test flow params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *SaveTestFlowParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the save test flow params
func (o *SaveTestFlowParams) WithTimeout(timeout time.Duration) *SaveTestFlowParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the save test flow params
func (o *SaveTestFlowParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the save test flow params
func (o *SaveTestFlowParams) WithContext(ctx context.Context) *SaveTestFlowParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the save test flow params
func (o *SaveTestFlowParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the save test flow params
func (o *SaveTestFlowParams) WithHTTPClient(client *http.Client) *SaveTestFlowParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the save test flow params
func (o *SaveTestFlowParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithTestflow adds the testflow to the save test flow params
func (o *SaveTestFlowParams) WithTestflow(testflow *models.TestFlow) *SaveTestFlowParams {
	o.SetTestflow(testflow)
	return o
}

// SetTestflow adds the testflow to the save test flow params
func (o *SaveTestFlowParams) SetTestflow(testflow *models.TestFlow) {
	o.Testflow = testflow
}

// WriteToRequest writes these params to a swagger request
func (o *SaveTestFlowParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Testflow != nil {
		if err := r.SetBodyParam(o.Testflow); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
