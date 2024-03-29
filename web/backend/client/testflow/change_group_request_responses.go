// Code generated by go-swagger; DO NOT EDIT.

package testflow

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// ChangeGroupRequestReader is a Reader for the ChangeGroupRequest structure.
type ChangeGroupRequestReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ChangeGroupRequestReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewChangeGroupRequestOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewChangeGroupRequestServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[POST /changegroup] changeGroupRequest", response, response.Code())
	}
}

// NewChangeGroupRequestOK creates a ChangeGroupRequestOK with default headers values
func NewChangeGroupRequestOK() *ChangeGroupRequestOK {
	return &ChangeGroupRequestOK{}
}

/*
ChangeGroupRequestOK describes a response with status code 200, with default header values.

ChangeGroupRequestOK change group request o k
*/
type ChangeGroupRequestOK struct {
}

// IsSuccess returns true when this change group request o k response has a 2xx status code
func (o *ChangeGroupRequestOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this change group request o k response has a 3xx status code
func (o *ChangeGroupRequestOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this change group request o k response has a 4xx status code
func (o *ChangeGroupRequestOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this change group request o k response has a 5xx status code
func (o *ChangeGroupRequestOK) IsServerError() bool {
	return false
}

// IsCode returns true when this change group request o k response a status code equal to that given
func (o *ChangeGroupRequestOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the change group request o k response
func (o *ChangeGroupRequestOK) Code() int {
	return 200
}

func (o *ChangeGroupRequestOK) Error() string {
	return fmt.Sprintf("[POST /changegroup][%d] changeGroupRequestOK ", 200)
}

func (o *ChangeGroupRequestOK) String() string {
	return fmt.Sprintf("[POST /changegroup][%d] changeGroupRequestOK ", 200)
}

func (o *ChangeGroupRequestOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewChangeGroupRequestServiceUnavailable creates a ChangeGroupRequestServiceUnavailable with default headers values
func NewChangeGroupRequestServiceUnavailable() *ChangeGroupRequestServiceUnavailable {
	return &ChangeGroupRequestServiceUnavailable{}
}

/*
ChangeGroupRequestServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type ChangeGroupRequestServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this change group request service unavailable response has a 2xx status code
func (o *ChangeGroupRequestServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this change group request service unavailable response has a 3xx status code
func (o *ChangeGroupRequestServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this change group request service unavailable response has a 4xx status code
func (o *ChangeGroupRequestServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this change group request service unavailable response has a 5xx status code
func (o *ChangeGroupRequestServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this change group request service unavailable response a status code equal to that given
func (o *ChangeGroupRequestServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the change group request service unavailable response
func (o *ChangeGroupRequestServiceUnavailable) Code() int {
	return 503
}

func (o *ChangeGroupRequestServiceUnavailable) Error() string {
	return fmt.Sprintf("[POST /changegroup][%d] changeGroupRequestServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ChangeGroupRequestServiceUnavailable) String() string {
	return fmt.Sprintf("[POST /changegroup][%d] changeGroupRequestServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ChangeGroupRequestServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ChangeGroupRequestServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
