// Code generated by go-swagger; DO NOT EDIT.

package failed_tasks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// ListFailedTasksReqReader is a Reader for the ListFailedTasksReq structure.
type ListFailedTasksReqReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListFailedTasksReqReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListFailedTasksReqOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewListFailedTasksReqInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /failed-tasks] listFailedTasksReq", response, response.Code())
	}
}

// NewListFailedTasksReqOK creates a ListFailedTasksReqOK with default headers values
func NewListFailedTasksReqOK() *ListFailedTasksReqOK {
	return &ListFailedTasksReqOK{}
}

/*
ListFailedTasksReqOK describes a response with status code 200, with default header values.

	//todo fix correctstruct
*/
type ListFailedTasksReqOK struct {
	Payload models.MyString
}

// IsSuccess returns true when this list failed tasks req o k response has a 2xx status code
func (o *ListFailedTasksReqOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list failed tasks req o k response has a 3xx status code
func (o *ListFailedTasksReqOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list failed tasks req o k response has a 4xx status code
func (o *ListFailedTasksReqOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list failed tasks req o k response has a 5xx status code
func (o *ListFailedTasksReqOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list failed tasks req o k response a status code equal to that given
func (o *ListFailedTasksReqOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list failed tasks req o k response
func (o *ListFailedTasksReqOK) Code() int {
	return 200
}

func (o *ListFailedTasksReqOK) Error() string {
	return fmt.Sprintf("[GET /failed-tasks][%d] listFailedTasksReqOK  %+v", 200, o.Payload)
}

func (o *ListFailedTasksReqOK) String() string {
	return fmt.Sprintf("[GET /failed-tasks][%d] listFailedTasksReqOK  %+v", 200, o.Payload)
}

func (o *ListFailedTasksReqOK) GetPayload() models.MyString {
	return o.Payload
}

func (o *ListFailedTasksReqOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListFailedTasksReqInternalServerError creates a ListFailedTasksReqInternalServerError with default headers values
func NewListFailedTasksReqInternalServerError() *ListFailedTasksReqInternalServerError {
	return &ListFailedTasksReqInternalServerError{}
}

/*
ListFailedTasksReqInternalServerError describes a response with status code 500, with default header values.

apiError
*/
type ListFailedTasksReqInternalServerError struct {
	Payload *models.APIError
}

// IsSuccess returns true when this list failed tasks req internal server error response has a 2xx status code
func (o *ListFailedTasksReqInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list failed tasks req internal server error response has a 3xx status code
func (o *ListFailedTasksReqInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list failed tasks req internal server error response has a 4xx status code
func (o *ListFailedTasksReqInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list failed tasks req internal server error response has a 5xx status code
func (o *ListFailedTasksReqInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list failed tasks req internal server error response a status code equal to that given
func (o *ListFailedTasksReqInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the list failed tasks req internal server error response
func (o *ListFailedTasksReqInternalServerError) Code() int {
	return 500
}

func (o *ListFailedTasksReqInternalServerError) Error() string {
	return fmt.Sprintf("[GET /failed-tasks][%d] listFailedTasksReqInternalServerError  %+v", 500, o.Payload)
}

func (o *ListFailedTasksReqInternalServerError) String() string {
	return fmt.Sprintf("[GET /failed-tasks][%d] listFailedTasksReqInternalServerError  %+v", 500, o.Payload)
}

func (o *ListFailedTasksReqInternalServerError) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ListFailedTasksReqInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
