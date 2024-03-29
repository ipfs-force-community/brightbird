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

// PassRateTrendsReader is a Reader for the PassRateTrends structure.
type PassRateTrendsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PassRateTrendsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPassRateTrendsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewPassRateTrendsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /pass-rate-trends] passRateTrends", response, response.Code())
	}
}

// NewPassRateTrendsOK creates a PassRateTrendsOK with default headers values
func NewPassRateTrendsOK() *PassRateTrendsOK {
	return &PassRateTrendsOK{}
}

/*
PassRateTrendsOK describes a response with status code 200, with default header values.

	//todo fix correctstruct
*/
type PassRateTrendsOK struct {
	Payload models.MyString
}

// IsSuccess returns true when this pass rate trends o k response has a 2xx status code
func (o *PassRateTrendsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this pass rate trends o k response has a 3xx status code
func (o *PassRateTrendsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this pass rate trends o k response has a 4xx status code
func (o *PassRateTrendsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this pass rate trends o k response has a 5xx status code
func (o *PassRateTrendsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this pass rate trends o k response a status code equal to that given
func (o *PassRateTrendsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the pass rate trends o k response
func (o *PassRateTrendsOK) Code() int {
	return 200
}

func (o *PassRateTrendsOK) Error() string {
	return fmt.Sprintf("[GET /pass-rate-trends][%d] passRateTrendsOK  %+v", 200, o.Payload)
}

func (o *PassRateTrendsOK) String() string {
	return fmt.Sprintf("[GET /pass-rate-trends][%d] passRateTrendsOK  %+v", 200, o.Payload)
}

func (o *PassRateTrendsOK) GetPayload() models.MyString {
	return o.Payload
}

func (o *PassRateTrendsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPassRateTrendsInternalServerError creates a PassRateTrendsInternalServerError with default headers values
func NewPassRateTrendsInternalServerError() *PassRateTrendsInternalServerError {
	return &PassRateTrendsInternalServerError{}
}

/*
PassRateTrendsInternalServerError describes a response with status code 500, with default header values.

apiError
*/
type PassRateTrendsInternalServerError struct {
	Payload *models.APIError
}

// IsSuccess returns true when this pass rate trends internal server error response has a 2xx status code
func (o *PassRateTrendsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this pass rate trends internal server error response has a 3xx status code
func (o *PassRateTrendsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this pass rate trends internal server error response has a 4xx status code
func (o *PassRateTrendsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this pass rate trends internal server error response has a 5xx status code
func (o *PassRateTrendsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this pass rate trends internal server error response a status code equal to that given
func (o *PassRateTrendsInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the pass rate trends internal server error response
func (o *PassRateTrendsInternalServerError) Code() int {
	return 500
}

func (o *PassRateTrendsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /pass-rate-trends][%d] passRateTrendsInternalServerError  %+v", 500, o.Payload)
}

func (o *PassRateTrendsInternalServerError) String() string {
	return fmt.Sprintf("[GET /pass-rate-trends][%d] passRateTrendsInternalServerError  %+v", 500, o.Payload)
}

func (o *PassRateTrendsInternalServerError) GetPayload() *models.APIError {
	return o.Payload
}

func (o *PassRateTrendsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
