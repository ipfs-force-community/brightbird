// Code generated by go-swagger; DO NOT EDIT.

package group

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// GetGroupByIDReader is a Reader for the GetGroupByID structure.
type GetGroupByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetGroupByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetGroupByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewGetGroupByIDServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /group/{id}] getGroupById", response, response.Code())
	}
}

// NewGetGroupByIDOK creates a GetGroupByIDOK with default headers values
func NewGetGroupByIDOK() *GetGroupByIDOK {
	return &GetGroupByIDOK{}
}

/*
GetGroupByIDOK describes a response with status code 200, with default header values.

groupResp
*/
type GetGroupByIDOK struct {
	Payload *models.GroupResp
}

// IsSuccess returns true when this get group by Id o k response has a 2xx status code
func (o *GetGroupByIDOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get group by Id o k response has a 3xx status code
func (o *GetGroupByIDOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get group by Id o k response has a 4xx status code
func (o *GetGroupByIDOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get group by Id o k response has a 5xx status code
func (o *GetGroupByIDOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get group by Id o k response a status code equal to that given
func (o *GetGroupByIDOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get group by Id o k response
func (o *GetGroupByIDOK) Code() int {
	return 200
}

func (o *GetGroupByIDOK) Error() string {
	return fmt.Sprintf("[GET /group/{id}][%d] getGroupByIdOK  %+v", 200, o.Payload)
}

func (o *GetGroupByIDOK) String() string {
	return fmt.Sprintf("[GET /group/{id}][%d] getGroupByIdOK  %+v", 200, o.Payload)
}

func (o *GetGroupByIDOK) GetPayload() *models.GroupResp {
	return o.Payload
}

func (o *GetGroupByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GroupResp)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetGroupByIDServiceUnavailable creates a GetGroupByIDServiceUnavailable with default headers values
func NewGetGroupByIDServiceUnavailable() *GetGroupByIDServiceUnavailable {
	return &GetGroupByIDServiceUnavailable{}
}

/*
GetGroupByIDServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type GetGroupByIDServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this get group by Id service unavailable response has a 2xx status code
func (o *GetGroupByIDServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get group by Id service unavailable response has a 3xx status code
func (o *GetGroupByIDServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get group by Id service unavailable response has a 4xx status code
func (o *GetGroupByIDServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this get group by Id service unavailable response has a 5xx status code
func (o *GetGroupByIDServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this get group by Id service unavailable response a status code equal to that given
func (o *GetGroupByIDServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the get group by Id service unavailable response
func (o *GetGroupByIDServiceUnavailable) Code() int {
	return 503
}

func (o *GetGroupByIDServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /group/{id}][%d] getGroupByIdServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetGroupByIDServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /group/{id}][%d] getGroupByIdServiceUnavailable  %+v", 503, o.Payload)
}

func (o *GetGroupByIDServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *GetGroupByIDServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
