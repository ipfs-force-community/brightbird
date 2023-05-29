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
)

// NewListLogsInPodParams creates a new ListLogsInPodParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListLogsInPodParams() *ListLogsInPodParams {
	return &ListLogsInPodParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListLogsInPodParamsWithTimeout creates a new ListLogsInPodParams object
// with the ability to set a timeout on a request.
func NewListLogsInPodParamsWithTimeout(timeout time.Duration) *ListLogsInPodParams {
	return &ListLogsInPodParams{
		timeout: timeout,
	}
}

// NewListLogsInPodParamsWithContext creates a new ListLogsInPodParams object
// with the ability to set a context for a request.
func NewListLogsInPodParamsWithContext(ctx context.Context) *ListLogsInPodParams {
	return &ListLogsInPodParams{
		Context: ctx,
	}
}

// NewListLogsInPodParamsWithHTTPClient creates a new ListLogsInPodParams object
// with the ability to set a custom HTTPClient for a request.
func NewListLogsInPodParamsWithHTTPClient(client *http.Client) *ListLogsInPodParams {
	return &ListLogsInPodParams{
		HTTPClient: client,
	}
}

/*
ListLogsInPodParams contains all the parameters to send to the API endpoint

	for the list logs in pod operation.

	Typically these are written to a http.Request.
*/
type ListLogsInPodParams struct {

	/* PodName.

	   pod name
	*/
	PodName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list logs in pod params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListLogsInPodParams) WithDefaults() *ListLogsInPodParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list logs in pod params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListLogsInPodParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list logs in pod params
func (o *ListLogsInPodParams) WithTimeout(timeout time.Duration) *ListLogsInPodParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list logs in pod params
func (o *ListLogsInPodParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list logs in pod params
func (o *ListLogsInPodParams) WithContext(ctx context.Context) *ListLogsInPodParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list logs in pod params
func (o *ListLogsInPodParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list logs in pod params
func (o *ListLogsInPodParams) WithHTTPClient(client *http.Client) *ListLogsInPodParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list logs in pod params
func (o *ListLogsInPodParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPodName adds the podName to the list logs in pod params
func (o *ListLogsInPodParams) WithPodName(podName string) *ListLogsInPodParams {
	o.SetPodName(podName)
	return o
}

// SetPodName adds the podName to the list logs in pod params
func (o *ListLogsInPodParams) SetPodName(podName string) {
	o.PodName = podName
}

// WriteToRequest writes these params to a swagger request
func (o *ListLogsInPodParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param podName
	if err := r.SetPathParam("podName", o.PodName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
