// Code generated by go-swagger; DO NOT EDIT.

package log

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
	"github.com/go-openapi/swag"
)

// NewListPodsReqParams creates a new ListPodsReqParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListPodsReqParams() *ListPodsReqParams {
	return &ListPodsReqParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListPodsReqParamsWithTimeout creates a new ListPodsReqParams object
// with the ability to set a timeout on a request.
func NewListPodsReqParamsWithTimeout(timeout time.Duration) *ListPodsReqParams {
	return &ListPodsReqParams{
		timeout: timeout,
	}
}

// NewListPodsReqParamsWithContext creates a new ListPodsReqParams object
// with the ability to set a context for a request.
func NewListPodsReqParamsWithContext(ctx context.Context) *ListPodsReqParams {
	return &ListPodsReqParams{
		Context: ctx,
	}
}

// NewListPodsReqParamsWithHTTPClient creates a new ListPodsReqParams object
// with the ability to set a custom HTTPClient for a request.
func NewListPodsReqParamsWithHTTPClient(client *http.Client) *ListPodsReqParams {
	return &ListPodsReqParams{
		HTTPClient: client,
	}
}

/*
ListPodsReqParams contains all the parameters to send to the API endpoint

	for the list pods req operation.

	Typically these are written to a http.Request.
*/
type ListPodsReqParams struct {

	/* RetryTime.

	   retrytime of task

	   Format: int64
	*/
	RetryTime int64

	/* TestID.

	   testid of task
	*/
	TestID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list pods req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListPodsReqParams) WithDefaults() *ListPodsReqParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list pods req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListPodsReqParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list pods req params
func (o *ListPodsReqParams) WithTimeout(timeout time.Duration) *ListPodsReqParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list pods req params
func (o *ListPodsReqParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list pods req params
func (o *ListPodsReqParams) WithContext(ctx context.Context) *ListPodsReqParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list pods req params
func (o *ListPodsReqParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list pods req params
func (o *ListPodsReqParams) WithHTTPClient(client *http.Client) *ListPodsReqParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list pods req params
func (o *ListPodsReqParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithRetryTime adds the retryTime to the list pods req params
func (o *ListPodsReqParams) WithRetryTime(retryTime int64) *ListPodsReqParams {
	o.SetRetryTime(retryTime)
	return o
}

// SetRetryTime adds the retryTime to the list pods req params
func (o *ListPodsReqParams) SetRetryTime(retryTime int64) {
	o.RetryTime = retryTime
}

// WithTestID adds the testID to the list pods req params
func (o *ListPodsReqParams) WithTestID(testID string) *ListPodsReqParams {
	o.SetTestID(testID)
	return o
}

// SetTestID adds the testId to the list pods req params
func (o *ListPodsReqParams) SetTestID(testID string) {
	o.TestID = testID
}

// WriteToRequest writes these params to a swagger request
func (o *ListPodsReqParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param retryTime
	qrRetryTime := o.RetryTime
	qRetryTime := swag.FormatInt64(qrRetryTime)
	if qRetryTime != "" {

		if err := r.SetQueryParam("retryTime", qRetryTime); err != nil {
			return err
		}
	}

	// query param testID
	qrTestID := o.TestID
	qTestID := qrTestID
	if qTestID != "" {

		if err := r.SetQueryParam("testID", qTestID); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
