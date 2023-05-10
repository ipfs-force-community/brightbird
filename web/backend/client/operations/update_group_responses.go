// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/hunjixin/brightbird/models"
)

// UpdateGroupReader is a Reader for the UpdateGroup structure.
type UpdateGroupReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateGroupReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateGroupOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewUpdateGroupServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewUpdateGroupOK creates a UpdateGroupOK with default headers values
func NewUpdateGroupOK() *UpdateGroupOK {
	return &UpdateGroupOK{}
}

/*
UpdateGroupOK describes a response with status code 200, with default header values.

UpdateGroupOK update group o k
*/
type UpdateGroupOK struct {
}

// IsSuccess returns true when this update group o k response has a 2xx status code
func (o *UpdateGroupOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this update group o k response has a 3xx status code
func (o *UpdateGroupOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update group o k response has a 4xx status code
func (o *UpdateGroupOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this update group o k response has a 5xx status code
func (o *UpdateGroupOK) IsServerError() bool {
	return false
}

// IsCode returns true when this update group o k response a status code equal to that given
func (o *UpdateGroupOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the update group o k response
func (o *UpdateGroupOK) Code() int {
	return 200
}

func (o *UpdateGroupOK) Error() string {
	return fmt.Sprintf("[POST /group/{id}][%d] updateGroupOK ", 200)
}

func (o *UpdateGroupOK) String() string {
	return fmt.Sprintf("[POST /group/{id}][%d] updateGroupOK ", 200)
}

func (o *UpdateGroupOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUpdateGroupServiceUnavailable creates a UpdateGroupServiceUnavailable with default headers values
func NewUpdateGroupServiceUnavailable() *UpdateGroupServiceUnavailable {
	return &UpdateGroupServiceUnavailable{}
}

/*
UpdateGroupServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type UpdateGroupServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this update group service unavailable response has a 2xx status code
func (o *UpdateGroupServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update group service unavailable response has a 3xx status code
func (o *UpdateGroupServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update group service unavailable response has a 4xx status code
func (o *UpdateGroupServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this update group service unavailable response has a 5xx status code
func (o *UpdateGroupServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this update group service unavailable response a status code equal to that given
func (o *UpdateGroupServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the update group service unavailable response
func (o *UpdateGroupServiceUnavailable) Code() int {
	return 503
}

func (o *UpdateGroupServiceUnavailable) Error() string {
	return fmt.Sprintf("[POST /group/{id}][%d] updateGroupServiceUnavailable  %+v", 503, o.Payload)
}

func (o *UpdateGroupServiceUnavailable) String() string {
	return fmt.Sprintf("[POST /group/{id}][%d] updateGroupServiceUnavailable  %+v", 503, o.Payload)
}

func (o *UpdateGroupServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *UpdateGroupServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
