// Code generated by go-swagger; DO NOT EDIT.

package version

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// GetVersionReader is a Reader for the GetVersion structure.
type GetVersionReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetVersionReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetVersionOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewGetVersionServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /version] getVersion", response, response.Code())
	}
}

// NewGetVersionOK creates a GetVersionOK with default headers values
func NewGetVersionOK() *GetVersionOK {
	return &GetVersionOK{}
}

/*
GetVersionOK describes a response with status code 200, with default header values.

myString
*/
type GetVersionOK struct {
	Payload models.MyString
}

// IsSuccess returns true when this get version o k response has a 2xx status code
func (o *GetVersionOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get version o k response has a 3xx status code
func (o *GetVersionOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get version o k response has a 4xx status code
func (o *GetVersionOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get version o k response has a 5xx status code
func (o *GetVersionOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get version o k response a status code equal to that given
func (o *GetVersionOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get version o k response
func (o *GetVersionOK) Code() int {
	return 200
}

func (o *GetVersionOK) Error() string {
	return fmt.Sprintf("[GET /version][%d] getVersionOK  %+v", 200, o.Payload)
}

func (o *GetVersionOK) String() string {
	return fmt.Sprintf("[GET /version][%d] getVersionOK  %+v", 200, o.Payload)
}

func (o *GetVersionOK) GetPayload() models.MyString {
	return o.Payload
}

func (o *GetVersionOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVersionServiceUnavailable creates a GetVersionServiceUnavailable with default headers values
func NewGetVersionServiceUnavailable() *GetVersionServiceUnavailable {
	return &GetVersionServiceUnavailable{}
}

/*
GetVersionServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type GetVersionServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this get version service unavailable response has a 2xx status code
func (o *GetVersionServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get version service unavailable response has a 3xx status code
func (o *GetVersionServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get version service unavailable response has a 4xx status code
func (o *GetVersionServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this get version service unavailable response has a 5xx status code
func (o *GetVersionServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this get version service unavailable response a status code equal to that given
func (o *GetVersionServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the get version service unavailable response
func (o *GetVersionServiceUnavailable) Code() int {
	return 503
}

func (o *GetVersionServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /version][%d] getVersionServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetVersionServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /version][%d] getVersionServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetVersionServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *GetVersionServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
