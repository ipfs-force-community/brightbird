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
)

// NewCountTestFlowRequestParams creates a new CountTestFlowRequestParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCountTestFlowRequestParams() *CountTestFlowRequestParams {
	return &CountTestFlowRequestParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCountTestFlowRequestParamsWithTimeout creates a new CountTestFlowRequestParams object
// with the ability to set a timeout on a request.
func NewCountTestFlowRequestParamsWithTimeout(timeout time.Duration) *CountTestFlowRequestParams {
	return &CountTestFlowRequestParams{
		timeout: timeout,
	}
}

// NewCountTestFlowRequestParamsWithContext creates a new CountTestFlowRequestParams object
// with the ability to set a context for a request.
func NewCountTestFlowRequestParamsWithContext(ctx context.Context) *CountTestFlowRequestParams {
	return &CountTestFlowRequestParams{
		Context: ctx,
	}
}

// NewCountTestFlowRequestParamsWithHTTPClient creates a new CountTestFlowRequestParams object
// with the ability to set a custom HTTPClient for a request.
func NewCountTestFlowRequestParamsWithHTTPClient(client *http.Client) *CountTestFlowRequestParams {
	return &CountTestFlowRequestParams{
		HTTPClient: client,
	}
}

/*
CountTestFlowRequestParams contains all the parameters to send to the API endpoint

	for the count test flow request operation.

	Typically these are written to a http.Request.
*/
type CountTestFlowRequestParams struct {

	/* GroupID.

	   id of group
	*/
	GroupID *string

	/* Name.

	   name of testflow
	*/
	Name *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the count test flow request params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CountTestFlowRequestParams) WithDefaults() *CountTestFlowRequestParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the count test flow request params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CountTestFlowRequestParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the count test flow request params
func (o *CountTestFlowRequestParams) WithTimeout(timeout time.Duration) *CountTestFlowRequestParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the count test flow request params
func (o *CountTestFlowRequestParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the count test flow request params
func (o *CountTestFlowRequestParams) WithContext(ctx context.Context) *CountTestFlowRequestParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the count test flow request params
func (o *CountTestFlowRequestParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the count test flow request params
func (o *CountTestFlowRequestParams) WithHTTPClient(client *http.Client) *CountTestFlowRequestParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the count test flow request params
func (o *CountTestFlowRequestParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithGroupID adds the groupID to the count test flow request params
func (o *CountTestFlowRequestParams) WithGroupID(groupID *string) *CountTestFlowRequestParams {
	o.SetGroupID(groupID)
	return o
}

// SetGroupID adds the groupId to the count test flow request params
func (o *CountTestFlowRequestParams) SetGroupID(groupID *string) {
	o.GroupID = groupID
}

// WithName adds the name to the count test flow request params
func (o *CountTestFlowRequestParams) WithName(name *string) *CountTestFlowRequestParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the count test flow request params
func (o *CountTestFlowRequestParams) SetName(name *string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *CountTestFlowRequestParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.GroupID != nil {

		// query param groupId
		var qrGroupID string

		if o.GroupID != nil {
			qrGroupID = *o.GroupID
		}
		qGroupID := qrGroupID
		if qGroupID != "" {

			if err := r.SetQueryParam("groupId", qGroupID); err != nil {
				return err
			}
		}
	}

	if o.Name != nil {

		// query param name
		var qrName string

		if o.Name != nil {
			qrName = *o.Name
		}
		qName := qrName
		if qName != "" {

			if err := r.SetQueryParam("name", qName); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}