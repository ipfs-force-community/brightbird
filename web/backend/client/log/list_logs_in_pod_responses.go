// Code generated by go-swagger; DO NOT EDIT.

package log

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/hunjixin/brightbird/models"
)

// ListLogsInPodReader is a Reader for the ListLogsInPod structure.
type ListLogsInPodReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListLogsInPodReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListLogsInPodOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewListLogsInPodServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListLogsInPodOK creates a ListLogsInPodOK with default headers values
func NewListLogsInPodOK() *ListLogsInPodOK {
	return &ListLogsInPodOK{}
}

/*
ListLogsInPodOK describes a response with status code 200, with default header values.

logResp
*/
type ListLogsInPodOK struct {
	Payload *models.LogResp
}

// IsSuccess returns true when this list logs in pod o k response has a 2xx status code
func (o *ListLogsInPodOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list logs in pod o k response has a 3xx status code
func (o *ListLogsInPodOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list logs in pod o k response has a 4xx status code
func (o *ListLogsInPodOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list logs in pod o k response has a 5xx status code
func (o *ListLogsInPodOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list logs in pod o k response a status code equal to that given
func (o *ListLogsInPodOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list logs in pod o k response
func (o *ListLogsInPodOK) Code() int {
	return 200
}

func (o *ListLogsInPodOK) Error() string {
	return fmt.Sprintf("[GET /logs/{podName}][%d] listLogsInPodOK  %+v", 200, o.Payload)
}

func (o *ListLogsInPodOK) String() string {
	return fmt.Sprintf("[GET /logs/{podName}][%d] listLogsInPodOK  %+v", 200, o.Payload)
}

func (o *ListLogsInPodOK) GetPayload() *models.LogResp {
	return o.Payload
}

func (o *ListLogsInPodOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.LogResp)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListLogsInPodServiceUnavailable creates a ListLogsInPodServiceUnavailable with default headers values
func NewListLogsInPodServiceUnavailable() *ListLogsInPodServiceUnavailable {
	return &ListLogsInPodServiceUnavailable{}
}

/*
ListLogsInPodServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type ListLogsInPodServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this list logs in pod service unavailable response has a 2xx status code
func (o *ListLogsInPodServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list logs in pod service unavailable response has a 3xx status code
func (o *ListLogsInPodServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list logs in pod service unavailable response has a 4xx status code
func (o *ListLogsInPodServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this list logs in pod service unavailable response has a 5xx status code
func (o *ListLogsInPodServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this list logs in pod service unavailable response a status code equal to that given
func (o *ListLogsInPodServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the list logs in pod service unavailable response
func (o *ListLogsInPodServiceUnavailable) Code() int {
	return 503
}

func (o *ListLogsInPodServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /logs/{podName}][%d] listLogsInPodServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListLogsInPodServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /logs/{podName}][%d] listLogsInPodServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListLogsInPodServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ListLogsInPodServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
