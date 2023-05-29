// Code generated by go-swagger; DO NOT EDIT.

package plugin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/hunjixin/brightbird/models"
)

// GetPluginMainfestReader is a Reader for the GetPluginMainfest structure.
type GetPluginMainfestReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPluginMainfestReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPluginMainfestOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewGetPluginMainfestServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetPluginMainfestOK creates a GetPluginMainfestOK with default headers values
func NewGetPluginMainfestOK() *GetPluginMainfestOK {
	return &GetPluginMainfestOK{}
}

/*
GetPluginMainfestOK describes a response with status code 200, with default header values.

pluginInfo
*/
type GetPluginMainfestOK struct {
	Payload []*models.PluginInfo
}

// IsSuccess returns true when this get plugin mainfest o k response has a 2xx status code
func (o *GetPluginMainfestOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get plugin mainfest o k response has a 3xx status code
func (o *GetPluginMainfestOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get plugin mainfest o k response has a 4xx status code
func (o *GetPluginMainfestOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get plugin mainfest o k response has a 5xx status code
func (o *GetPluginMainfestOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get plugin mainfest o k response a status code equal to that given
func (o *GetPluginMainfestOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get plugin mainfest o k response
func (o *GetPluginMainfestOK) Code() int {
	return 200
}

func (o *GetPluginMainfestOK) Error() string {
	return fmt.Sprintf("[GET /plugin/mainfest][%d] getPluginMainfestOK  %+v", 200, o.Payload)
}

func (o *GetPluginMainfestOK) String() string {
	return fmt.Sprintf("[GET /plugin/mainfest][%d] getPluginMainfestOK  %+v", 200, o.Payload)
}

func (o *GetPluginMainfestOK) GetPayload() []*models.PluginInfo {
	return o.Payload
}

func (o *GetPluginMainfestOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPluginMainfestServiceUnavailable creates a GetPluginMainfestServiceUnavailable with default headers values
func NewGetPluginMainfestServiceUnavailable() *GetPluginMainfestServiceUnavailable {
	return &GetPluginMainfestServiceUnavailable{}
}

/*
GetPluginMainfestServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type GetPluginMainfestServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this get plugin mainfest service unavailable response has a 2xx status code
func (o *GetPluginMainfestServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get plugin mainfest service unavailable response has a 3xx status code
func (o *GetPluginMainfestServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get plugin mainfest service unavailable response has a 4xx status code
func (o *GetPluginMainfestServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this get plugin mainfest service unavailable response has a 5xx status code
func (o *GetPluginMainfestServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this get plugin mainfest service unavailable response a status code equal to that given
func (o *GetPluginMainfestServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the get plugin mainfest service unavailable response
func (o *GetPluginMainfestServiceUnavailable) Code() int {
	return 503
}

func (o *GetPluginMainfestServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /plugin/mainfest][%d] getPluginMainfestServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetPluginMainfestServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /plugin/mainfest][%d] getPluginMainfestServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetPluginMainfestServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *GetPluginMainfestServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
