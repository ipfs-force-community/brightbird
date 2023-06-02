// Code generated by go-swagger; DO NOT EDIT.

package task

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/hunjixin/brightbird/models"
)

// ListTasksReader is a Reader for the ListTasks structure.
type ListTasksReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListTasksReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListTasksOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewListTasksServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListTasksOK creates a ListTasksOK with default headers values
func NewListTasksOK() *ListTasksOK {
	return &ListTasksOK{}
}

/*
ListTasksOK describes a response with status code 200, with default header values.

listTasksResp
*/
type ListTasksOK struct {
	Payload *models.ListTasksResp
}

// IsSuccess returns true when this list tasks o k response has a 2xx status code
func (o *ListTasksOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list tasks o k response has a 3xx status code
func (o *ListTasksOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks o k response has a 4xx status code
func (o *ListTasksOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tasks o k response has a 5xx status code
func (o *ListTasksOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks o k response a status code equal to that given
func (o *ListTasksOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list tasks o k response
func (o *ListTasksOK) Code() int {
	return 200
}

func (o *ListTasksOK) Error() string {
	return fmt.Sprintf("[GET /task][%d] listTasksOK  %+v", 200, o.Payload)
}

func (o *ListTasksOK) String() string {
	return fmt.Sprintf("[GET /task][%d] listTasksOK  %+v", 200, o.Payload)
}

func (o *ListTasksOK) GetPayload() *models.ListTasksResp {
	return o.Payload
}

func (o *ListTasksOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ListTasksResp)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListTasksServiceUnavailable creates a ListTasksServiceUnavailable with default headers values
func NewListTasksServiceUnavailable() *ListTasksServiceUnavailable {
	return &ListTasksServiceUnavailable{}
}

/*
ListTasksServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type ListTasksServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this list tasks service unavailable response has a 2xx status code
func (o *ListTasksServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks service unavailable response has a 3xx status code
func (o *ListTasksServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks service unavailable response has a 4xx status code
func (o *ListTasksServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tasks service unavailable response has a 5xx status code
func (o *ListTasksServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this list tasks service unavailable response a status code equal to that given
func (o *ListTasksServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the list tasks service unavailable response
func (o *ListTasksServiceUnavailable) Code() int {
	return 503
}

func (o *ListTasksServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /task][%d] listTasksServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListTasksServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /task][%d] listTasksServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListTasksServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ListTasksServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}