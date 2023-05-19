// Code generated by go-swagger; DO NOT EDIT.

package operations

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

// NewUpdateGroupParams creates a new UpdateGroupParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewUpdateGroupParams() *UpdateGroupParams {
	return &UpdateGroupParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewUpdateGroupParamsWithTimeout creates a new UpdateGroupParams object
// with the ability to set a timeout on a request.
func NewUpdateGroupParamsWithTimeout(timeout time.Duration) *UpdateGroupParams {
	return &UpdateGroupParams{
		timeout: timeout,
	}
}

// NewUpdateGroupParamsWithContext creates a new UpdateGroupParams object
// with the ability to set a context for a request.
func NewUpdateGroupParamsWithContext(ctx context.Context) *UpdateGroupParams {
	return &UpdateGroupParams{
		Context: ctx,
	}
}

// NewUpdateGroupParamsWithHTTPClient creates a new UpdateGroupParams object
// with the ability to set a custom HTTPClient for a request.
func NewUpdateGroupParamsWithHTTPClient(client *http.Client) *UpdateGroupParams {
	return &UpdateGroupParams{
		HTTPClient: client,
	}
}

/*
UpdateGroupParams contains all the parameters to send to the API endpoint

	for the update group operation.

	Typically these are written to a http.Request.
*/
type UpdateGroupParams struct {

	/* Group.

	   update group request json
	*/
	Group *models.UpdateGroupRequest

	/* ID.

	   id of group
	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the update group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UpdateGroupParams) WithDefaults() *UpdateGroupParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the update group params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UpdateGroupParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the update group params
func (o *UpdateGroupParams) WithTimeout(timeout time.Duration) *UpdateGroupParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the update group params
func (o *UpdateGroupParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the update group params
func (o *UpdateGroupParams) WithContext(ctx context.Context) *UpdateGroupParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the update group params
func (o *UpdateGroupParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the update group params
func (o *UpdateGroupParams) WithHTTPClient(client *http.Client) *UpdateGroupParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the update group params
func (o *UpdateGroupParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithGroup adds the group to the update group params
func (o *UpdateGroupParams) WithGroup(group *models.UpdateGroupRequest) *UpdateGroupParams {
	o.SetGroup(group)
	return o
}

// SetGroup adds the group to the update group params
func (o *UpdateGroupParams) SetGroup(group *models.UpdateGroupRequest) {
	o.Group = group
}

// WithID adds the id to the update group params
func (o *UpdateGroupParams) WithID(id string) *UpdateGroupParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the update group params
func (o *UpdateGroupParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *UpdateGroupParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Group != nil {
		if err := r.SetBodyParam(o.Group); err != nil {
			return err
		}
	}

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
