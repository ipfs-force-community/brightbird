// Code generated by go-swagger; DO NOT EDIT.

package plugin

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

// NewMyOperationParams creates a new MyOperationParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewMyOperationParams() *MyOperationParams {
	return &MyOperationParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewMyOperationParamsWithTimeout creates a new MyOperationParams object
// with the ability to set a timeout on a request.
func NewMyOperationParamsWithTimeout(timeout time.Duration) *MyOperationParams {
	return &MyOperationParams{
		timeout: timeout,
	}
}

// NewMyOperationParamsWithContext creates a new MyOperationParams object
// with the ability to set a context for a request.
func NewMyOperationParamsWithContext(ctx context.Context) *MyOperationParams {
	return &MyOperationParams{
		Context: ctx,
	}
}

// NewMyOperationParamsWithHTTPClient creates a new MyOperationParams object
// with the ability to set a custom HTTPClient for a request.
func NewMyOperationParamsWithHTTPClient(client *http.Client) *MyOperationParams {
	return &MyOperationParams{
		HTTPClient: client,
	}
}

/*
MyOperationParams contains all the parameters to send to the API endpoint

	for the my operation operation.

	Typically these are written to a http.Request.
*/
type MyOperationParams struct {

	/* MyFormFile.

	   MyFormFile desc.
	*/
	MyFormFile runtime.NamedReadCloser

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the my operation params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MyOperationParams) WithDefaults() *MyOperationParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the my operation params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MyOperationParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the my operation params
func (o *MyOperationParams) WithTimeout(timeout time.Duration) *MyOperationParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the my operation params
func (o *MyOperationParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the my operation params
func (o *MyOperationParams) WithContext(ctx context.Context) *MyOperationParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the my operation params
func (o *MyOperationParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the my operation params
func (o *MyOperationParams) WithHTTPClient(client *http.Client) *MyOperationParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the my operation params
func (o *MyOperationParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithMyFormFile adds the myFormFile to the my operation params
func (o *MyOperationParams) WithMyFormFile(myFormFile runtime.NamedReadCloser) *MyOperationParams {
	o.SetMyFormFile(myFormFile)
	return o
}

// SetMyFormFile adds the myFormFile to the my operation params
func (o *MyOperationParams) SetMyFormFile(myFormFile runtime.NamedReadCloser) {
	o.MyFormFile = myFormFile
}

// WriteToRequest writes these params to a swagger request
func (o *MyOperationParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.MyFormFile != nil {

		if o.MyFormFile != nil {
			// form file param myFormFile
			if err := r.SetFileParam("myFormFile", o.MyFormFile); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
