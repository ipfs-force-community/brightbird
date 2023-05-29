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

// NewCountTestFlowsInGroupParams creates a new CountTestFlowsInGroupParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCountTestFlowsInGroupParams() *CountTestFlowsInGroupParams {
	return &CountTestFlowsInGroupParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCountTestFlowsInGroupParamsWithTimeout creates a new CountTestFlowsInGroupParams object
// with the ability to set a timeout on a request.
func NewCountTestFlowsInGroupParamsWithTimeout(timeout time.Duration) *CountTestFlowsInGroupParams {
	return &CountTestFlowsInGroupParams{
		timeout: timeout,
	}
}

// NewCountTestFlowsInGroupParamsWithContext creates a new CountTestFlowsInGroupParams object
// with the ability to set a context for a request.
func NewCountTestFlowsInGroupParamsWithContext(ctx context.Context) *CountTestFlowsInGroupParams {
	return &CountTestFlowsInGroupParams{
		Context: ctx,
	}
}

// NewCountTestFlowsInGroupParamsWithHTTPClient creates a new CountTestFlowsInGroupParams object
// with the ability to set a custom HTTPClient for a request.
func NewCountTestFlowsInGroupParamsWithHTTPClient(client *http.Client) *CountTestFlowsInGroupParams {
	return &CountTestFlowsInGroupParams{
		HTTPClient: client,
	}
}

/*
CountTestFlowsInGroupParams contains all the parameters to send to the API endpoint

	for the count test flows in group operation.

	Typically these are written to a http.Request.
*/
type CountTestFlowsInGroupParams struct {

	/* GroupID.

	   group id  of test flow
	*/
	GroupID *string

	/* Name.

	   name  of test flow
	*/
	Name *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the count test flows in group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CountTestFlowsInGroupParams) WithDefaults() *CountTestFlowsInGroupParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the count test flows in group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CountTestFlowsInGroupParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the count test flows in group params
func (o *CountTestFlowsInGroupParams) WithTimeout(timeout time.Duration) *CountTestFlowsInGroupParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the count test flows in group params
func (o *CountTestFlowsInGroupParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the count test flows in group params
func (o *CountTestFlowsInGroupParams) WithContext(ctx context.Context) *CountTestFlowsInGroupParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the count test flows in group params
func (o *CountTestFlowsInGroupParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the count test flows in group params
func (o *CountTestFlowsInGroupParams) WithHTTPClient(client *http.Client) *CountTestFlowsInGroupParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the count test flows in group params
func (o *CountTestFlowsInGroupParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithGroupID adds the groupID to the count test flows in group params
func (o *CountTestFlowsInGroupParams) WithGroupID(groupID *string) *CountTestFlowsInGroupParams {
	o.SetGroupID(groupID)
	return o
}

// SetGroupID adds the groupId to the count test flows in group params
func (o *CountTestFlowsInGroupParams) SetGroupID(groupID *string) {
	o.GroupID = groupID
}

// WithName adds the name to the count test flows in group params
func (o *CountTestFlowsInGroupParams) WithName(name *string) *CountTestFlowsInGroupParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the count test flows in group params
func (o *CountTestFlowsInGroupParams) SetName(name *string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *CountTestFlowsInGroupParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
