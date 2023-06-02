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
	"github.com/go-openapi/swag"
)

// NewListTasksReqParams creates a new ListTasksReqParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListTasksReqParams() *ListTasksReqParams {
	return &ListTasksReqParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListTasksReqParamsWithTimeout creates a new ListTasksReqParams object
// with the ability to set a timeout on a request.
func NewListTasksReqParamsWithTimeout(timeout time.Duration) *ListTasksReqParams {
	return &ListTasksReqParams{
		timeout: timeout,
	}
}

// NewListTasksReqParamsWithContext creates a new ListTasksReqParams object
// with the ability to set a context for a request.
func NewListTasksReqParamsWithContext(ctx context.Context) *ListTasksReqParams {
	return &ListTasksReqParams{
		Context: ctx,
	}
}

// NewListTasksReqParamsWithHTTPClient creates a new ListTasksReqParams object
// with the ability to set a custom HTTPClient for a request.
func NewListTasksReqParamsWithHTTPClient(client *http.Client) *ListTasksReqParams {
	return &ListTasksReqParams{
		HTTPClient: client,
	}
}

/*
ListTasksReqParams contains all the parameters to send to the API endpoint

	for the list tasks req operation.

	Typically these are written to a http.Request.
*/
type ListTasksReqParams struct {

	/* CreateTime.

	   createtime of task

	   Format: int64
	*/
	CreateTime *int64

	/* JobID.

	   id of job
	*/
	JobID *string

	/* PageNum.

	   pageNum

	   Format: int64
	*/
	PageNum int64

	/* PageSize.

	   pageSize

	   Format: int64
	*/
	PageSize int64

	/* State.

	   task state
	*/
	State []int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list tasks req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListTasksReqParams) WithDefaults() *ListTasksReqParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list tasks req params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListTasksReqParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list tasks req params
func (o *ListTasksReqParams) WithTimeout(timeout time.Duration) *ListTasksReqParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list tasks req params
func (o *ListTasksReqParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list tasks req params
func (o *ListTasksReqParams) WithContext(ctx context.Context) *ListTasksReqParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list tasks req params
func (o *ListTasksReqParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list tasks req params
func (o *ListTasksReqParams) WithHTTPClient(client *http.Client) *ListTasksReqParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list tasks req params
func (o *ListTasksReqParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithCreateTime adds the createTime to the list tasks req params
func (o *ListTasksReqParams) WithCreateTime(createTime *int64) *ListTasksReqParams {
	o.SetCreateTime(createTime)
	return o
}

// SetCreateTime adds the createTime to the list tasks req params
func (o *ListTasksReqParams) SetCreateTime(createTime *int64) {
	o.CreateTime = createTime
}

// WithJobID adds the jobID to the list tasks req params
func (o *ListTasksReqParams) WithJobID(jobID *string) *ListTasksReqParams {
	o.SetJobID(jobID)
	return o
}

// SetJobID adds the jobId to the list tasks req params
func (o *ListTasksReqParams) SetJobID(jobID *string) {
	o.JobID = jobID
}

// WithPageNum adds the pageNum to the list tasks req params
func (o *ListTasksReqParams) WithPageNum(pageNum int64) *ListTasksReqParams {
	o.SetPageNum(pageNum)
	return o
}

// SetPageNum adds the pageNum to the list tasks req params
func (o *ListTasksReqParams) SetPageNum(pageNum int64) {
	o.PageNum = pageNum
}

// WithPageSize adds the pageSize to the list tasks req params
func (o *ListTasksReqParams) WithPageSize(pageSize int64) *ListTasksReqParams {
	o.SetPageSize(pageSize)
	return o
}

// SetPageSize adds the pageSize to the list tasks req params
func (o *ListTasksReqParams) SetPageSize(pageSize int64) {
	o.PageSize = pageSize
}

// WithState adds the state to the list tasks req params
func (o *ListTasksReqParams) WithState(state []int64) *ListTasksReqParams {
	o.SetState(state)
	return o
}

// SetState adds the state to the list tasks req params
func (o *ListTasksReqParams) SetState(state []int64) {
	o.State = state
}

// WriteToRequest writes these params to a swagger request
func (o *ListTasksReqParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.CreateTime != nil {

		// query param createTime
		var qrCreateTime int64

		if o.CreateTime != nil {
			qrCreateTime = *o.CreateTime
		}
		qCreateTime := swag.FormatInt64(qrCreateTime)
		if qCreateTime != "" {

			if err := r.SetQueryParam("createTime", qCreateTime); err != nil {
				return err
			}
		}
	}

	if o.JobID != nil {

		// query param jobId
		var qrJobID string

		if o.JobID != nil {
			qrJobID = *o.JobID
		}
		qJobID := qrJobID
		if qJobID != "" {

			if err := r.SetQueryParam("jobId", qJobID); err != nil {
				return err
			}
		}
	}

	// query param pageNum
	qrPageNum := o.PageNum
	qPageNum := swag.FormatInt64(qrPageNum)
	if qPageNum != "" {

		if err := r.SetQueryParam("pageNum", qPageNum); err != nil {
			return err
		}
	}

	// query param pageSize
	qrPageSize := o.PageSize
	qPageSize := swag.FormatInt64(qrPageSize)
	if qPageSize != "" {

		if err := r.SetQueryParam("pageSize", qPageSize); err != nil {
			return err
		}
	}

	if o.State != nil {

		// binding items for state
		joinedState := o.bindParamState(reg)

		// query array param state
		if err := r.SetQueryParam("state", joinedState...); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindParamListTasksReq binds the parameter state
func (o *ListTasksReqParams) bindParamState(formats strfmt.Registry) []string {
	stateIR := o.State

	var stateIC []string
	for _, stateIIR := range stateIR { // explode []int64

		stateIIV := swag.FormatInt64(stateIIR) // int64 as string
		stateIC = append(stateIC, stateIIV)
	}

	// items.CollectionFormat: ""
	stateIS := swag.JoinByFormat(stateIC, "")

	return stateIS
}