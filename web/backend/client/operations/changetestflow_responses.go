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

// ChangetestflowReader is a Reader for the Changetestflow structure.
type ChangetestflowReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ChangetestflowReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewChangetestflowOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewChangetestflowServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewChangetestflowOK creates a ChangetestflowOK with default headers values
func NewChangetestflowOK() *ChangetestflowOK {
	return &ChangetestflowOK{}
}

/*
ChangetestflowOK describes a response with status code 200, with default header values.

ChangetestflowOK changetestflow o k
*/
type ChangetestflowOK struct {
}

// IsSuccess returns true when this changetestflow o k response has a 2xx status code
func (o *ChangetestflowOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this changetestflow o k response has a 3xx status code
func (o *ChangetestflowOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this changetestflow o k response has a 4xx status code
func (o *ChangetestflowOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this changetestflow o k response has a 5xx status code
func (o *ChangetestflowOK) IsServerError() bool {
	return false
}

// IsCode returns true when this changetestflow o k response a status code equal to that given
func (o *ChangetestflowOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the changetestflow o k response
func (o *ChangetestflowOK) Code() int {
	return 200
}

func (o *ChangetestflowOK) Error() string {
	return fmt.Sprintf("[POST /changegroup][%d] changetestflowOK ", 200)
}

func (o *ChangetestflowOK) String() string {
	return fmt.Sprintf("[POST /changegroup][%d] changetestflowOK ", 200)
}

func (o *ChangetestflowOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewChangetestflowServiceUnavailable creates a ChangetestflowServiceUnavailable with default headers values
func NewChangetestflowServiceUnavailable() *ChangetestflowServiceUnavailable {
	return &ChangetestflowServiceUnavailable{}
}

/*
ChangetestflowServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type ChangetestflowServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this changetestflow service unavailable response has a 2xx status code
func (o *ChangetestflowServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this changetestflow service unavailable response has a 3xx status code
func (o *ChangetestflowServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this changetestflow service unavailable response has a 4xx status code
func (o *ChangetestflowServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this changetestflow service unavailable response has a 5xx status code
func (o *ChangetestflowServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this changetestflow service unavailable response a status code equal to that given
func (o *ChangetestflowServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the changetestflow service unavailable response
func (o *ChangetestflowServiceUnavailable) Code() int {
	return 503
}

func (o *ChangetestflowServiceUnavailable) Error() string {
	return fmt.Sprintf("[POST /changegroup][%d] changetestflowServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ChangetestflowServiceUnavailable) String() string {
	return fmt.Sprintf("[POST /changegroup][%d] changetestflowServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ChangetestflowServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ChangetestflowServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
