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
	"github.com/go-openapi/swag"
)

// NewListTestFlowsInGroupParams creates a new ListTestFlowsInGroupParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListTestFlowsInGroupParams() *ListTestFlowsInGroupParams {
	return &ListTestFlowsInGroupParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListTestFlowsInGroupParamsWithTimeout creates a new ListTestFlowsInGroupParams object
// with the ability to set a timeout on a request.
func NewListTestFlowsInGroupParamsWithTimeout(timeout time.Duration) *ListTestFlowsInGroupParams {
	return &ListTestFlowsInGroupParams{
		timeout: timeout,
	}
}

// NewListTestFlowsInGroupParamsWithContext creates a new ListTestFlowsInGroupParams object
// with the ability to set a context for a request.
func NewListTestFlowsInGroupParamsWithContext(ctx context.Context) *ListTestFlowsInGroupParams {
	return &ListTestFlowsInGroupParams{
		Context: ctx,
	}
}

// NewListTestFlowsInGroupParamsWithHTTPClient creates a new ListTestFlowsInGroupParams object
// with the ability to set a custom HTTPClient for a request.
func NewListTestFlowsInGroupParamsWithHTTPClient(client *http.Client) *ListTestFlowsInGroupParams {
	return &ListTestFlowsInGroupParams{
		HTTPClient: client,
	}
}

/*
ListTestFlowsInGroupParams contains all the parameters to send to the API endpoint

	for the list test flows in group operation.

	Typically these are written to a http.Request.
*/
type ListTestFlowsInGroupParams struct {

	/* GroupID.

	   group id  of test flow
	*/
	GroupID string

	/* PageNum.

	   page number  of test flow
	*/
	PageNum *int64

	/* PageSize.

	   page size  of test flow
	*/
	PageSize *int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list test flows in group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListTestFlowsInGroupParams) WithDefaults() *ListTestFlowsInGroupParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list test flows in group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListTestFlowsInGroupParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithTimeout(timeout time.Duration) *ListTestFlowsInGroupParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithContext(ctx context.Context) *ListTestFlowsInGroupParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithHTTPClient(client *http.Client) *ListTestFlowsInGroupParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithGroupID adds the groupID to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithGroupID(groupID string) *ListTestFlowsInGroupParams {
	o.SetGroupID(groupID)
	return o
}

// SetGroupID adds the groupId to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetGroupID(groupID string) {
	o.GroupID = groupID
}

// WithPageNum adds the pageNum to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithPageNum(pageNum *int64) *ListTestFlowsInGroupParams {
	o.SetPageNum(pageNum)
	return o
}

// SetPageNum adds the pageNum to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetPageNum(pageNum *int64) {
	o.PageNum = pageNum
}

// WithPageSize adds the pageSize to the list test flows in group params
func (o *ListTestFlowsInGroupParams) WithPageSize(pageSize *int64) *ListTestFlowsInGroupParams {
	o.SetPageSize(pageSize)
	return o
}

// SetPageSize adds the pageSize to the list test flows in group params
func (o *ListTestFlowsInGroupParams) SetPageSize(pageSize *int64) {
	o.PageSize = pageSize
}

// WriteToRequest writes these params to a swagger request
func (o *ListTestFlowsInGroupParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param groupId
	qrGroupID := o.GroupID
	qGroupID := qrGroupID
	if qGroupID != "" {

		if err := r.SetQueryParam("groupId", qGroupID); err != nil {
			return err
		}
	}

	if o.PageNum != nil {

		// query param pageNum
		var qrPageNum int64

		if o.PageNum != nil {
			qrPageNum = *o.PageNum
		}
		qPageNum := swag.FormatInt64(qrPageNum)
		if qPageNum != "" {

			if err := r.SetQueryParam("pageNum", qPageNum); err != nil {
				return err
			}
		}
	}

	if o.PageSize != nil {

		// query param pageSize
		var qrPageSize int64

		if o.PageSize != nil {
			qrPageSize = *o.PageSize
		}
		qPageSize := swag.FormatInt64(qrPageSize)
		if qPageSize != "" {

			if err := r.SetQueryParam("pageSize", qPageSize); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
