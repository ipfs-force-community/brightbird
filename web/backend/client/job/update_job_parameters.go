// Code generated by go-swagger; DO NOT EDIT.

package job

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

	"github.com/ipfs-force-community/brightbird/models"
)

// NewUpdateJobParams creates a new UpdateJobParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewUpdateJobParams() *UpdateJobParams {
	return &UpdateJobParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewUpdateJobParamsWithTimeout creates a new UpdateJobParams object
// with the ability to set a timeout on a request.
func NewUpdateJobParamsWithTimeout(timeout time.Duration) *UpdateJobParams {
	return &UpdateJobParams{
		timeout: timeout,
	}
}

// NewUpdateJobParamsWithContext creates a new UpdateJobParams object
// with the ability to set a context for a request.
func NewUpdateJobParamsWithContext(ctx context.Context) *UpdateJobParams {
	return &UpdateJobParams{
		Context: ctx,
	}
}

// NewUpdateJobParamsWithHTTPClient creates a new UpdateJobParams object
// with the ability to set a custom HTTPClient for a request.
func NewUpdateJobParamsWithHTTPClient(client *http.Client) *UpdateJobParams {
	return &UpdateJobParams{
		HTTPClient: client,
	}
}

/*
UpdateJobParams contains all the parameters to send to the API endpoint

	for the update job operation.

	Typically these are written to a http.Request.
*/
type UpdateJobParams struct {

	/* ID.

	   job id
	*/
	ID string

	/* UpdateJobParams.

	   job update params
	*/
	UpdateJobParams *models.UpdateJobRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the update job params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UpdateJobParams) WithDefaults() *UpdateJobParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the update job params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UpdateJobParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the update job params
func (o *UpdateJobParams) WithTimeout(timeout time.Duration) *UpdateJobParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the update job params
func (o *UpdateJobParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the update job params
func (o *UpdateJobParams) WithContext(ctx context.Context) *UpdateJobParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the update job params
func (o *UpdateJobParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the update job params
func (o *UpdateJobParams) WithHTTPClient(client *http.Client) *UpdateJobParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the update job params
func (o *UpdateJobParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the update job params
func (o *UpdateJobParams) WithID(id string) *UpdateJobParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the update job params
func (o *UpdateJobParams) SetID(id string) {
	o.ID = id
}

// WithUpdateJobParams adds the updateJobParams to the update job params
func (o *UpdateJobParams) WithUpdateJobParams(updateJobParams *models.UpdateJobRequest) *UpdateJobParams {
	o.SetUpdateJobParams(updateJobParams)
	return o
}

// SetUpdateJobParams adds the updateJobParams to the update job params
func (o *UpdateJobParams) SetUpdateJobParams(updateJobParams *models.UpdateJobRequest) {
	o.UpdateJobParams = updateJobParams
}

// WriteToRequest writes these params to a swagger request
func (o *UpdateJobParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}
	if o.UpdateJobParams != nil {
		if err := r.SetBodyParam(o.UpdateJobParams); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
