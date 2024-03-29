// Code generated by go-swagger; DO NOT EDIT.

package task

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// RetryTaskReqReader is a Reader for the RetryTaskReq structure.
type RetryTaskReqReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RetryTaskReqReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewRetryTaskReqOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewRetryTaskReqServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /task/retry] retryTaskReq", response, response.Code())
	}
}

// NewRetryTaskReqOK creates a RetryTaskReqOK with default headers values
func NewRetryTaskReqOK() *RetryTaskReqOK {
	return &RetryTaskReqOK{}
}

/*
RetryTaskReqOK describes a response with status code 200, with default header values.

int64Arr
*/
type RetryTaskReqOK struct {
	Payload models.Int64Array
}

// IsSuccess returns true when this retry task req o k response has a 2xx status code
func (o *RetryTaskReqOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this retry task req o k response has a 3xx status code
func (o *RetryTaskReqOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this retry task req o k response has a 4xx status code
func (o *RetryTaskReqOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this retry task req o k response has a 5xx status code
func (o *RetryTaskReqOK) IsServerError() bool {
	return false
}

// IsCode returns true when this retry task req o k response a status code equal to that given
func (o *RetryTaskReqOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the retry task req o k response
func (o *RetryTaskReqOK) Code() int {
	return 200
}

func (o *RetryTaskReqOK) Error() string {
	return fmt.Sprintf("[GET /task/retry][%d] retryTaskReqOK  %+v", 200, o.Payload)
}

func (o *RetryTaskReqOK) String() string {
	return fmt.Sprintf("[GET /task/retry][%d] retryTaskReqOK  %+v", 200, o.Payload)
}

func (o *RetryTaskReqOK) GetPayload() models.Int64Array {
	return o.Payload
}

func (o *RetryTaskReqOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRetryTaskReqServiceUnavailable creates a RetryTaskReqServiceUnavailable with default headers values
func NewRetryTaskReqServiceUnavailable() *RetryTaskReqServiceUnavailable {
	return &RetryTaskReqServiceUnavailable{}
}

/*
RetryTaskReqServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type RetryTaskReqServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this retry task req service unavailable response has a 2xx status code
func (o *RetryTaskReqServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this retry task req service unavailable response has a 3xx status code
func (o *RetryTaskReqServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this retry task req service unavailable response has a 4xx status code
func (o *RetryTaskReqServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this retry task req service unavailable response has a 5xx status code
func (o *RetryTaskReqServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this retry task req service unavailable response a status code equal to that given
func (o *RetryTaskReqServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the retry task req service unavailable response
func (o *RetryTaskReqServiceUnavailable) Code() int {
	return 503
}

func (o *RetryTaskReqServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /task/retry][%d] retryTaskReqServiceUnavailable  %+v", 503, o.Payload)
}

func (o *RetryTaskReqServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /task/retry][%d] retryTaskReqServiceUnavailable  %+v", 503, o.Payload)
}

func (o *RetryTaskReqServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *RetryTaskReqServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
