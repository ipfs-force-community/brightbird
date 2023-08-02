// Code generated by go-swagger; DO NOT EDIT.

package download

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ipfs-force-community/brightbird/models"
)

// DownloadParamsReader is a Reader for the DownloadParams structure.
type DownloadParamsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DownloadParamsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 500:
		result := NewDownloadParamsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /download] downloadParams", response, response.Code())
	}
}

// NewDownloadParamsInternalServerError creates a DownloadParamsInternalServerError with default headers values
func NewDownloadParamsInternalServerError() *DownloadParamsInternalServerError {
	return &DownloadParamsInternalServerError{}
}

/*
DownloadParamsInternalServerError describes a response with status code 500, with default header values.

apiError
*/
type DownloadParamsInternalServerError struct {
	Payload *models.APIError
}

// IsSuccess returns true when this download params internal server error response has a 2xx status code
func (o *DownloadParamsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this download params internal server error response has a 3xx status code
func (o *DownloadParamsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download params internal server error response has a 4xx status code
func (o *DownloadParamsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this download params internal server error response has a 5xx status code
func (o *DownloadParamsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this download params internal server error response a status code equal to that given
func (o *DownloadParamsInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the download params internal server error response
func (o *DownloadParamsInternalServerError) Code() int {
	return 500
}

func (o *DownloadParamsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /download][%d] downloadParamsInternalServerError  %+v", 500, o.Payload)
}

func (o *DownloadParamsInternalServerError) String() string {
	return fmt.Sprintf("[GET /download][%d] downloadParamsInternalServerError  %+v", 500, o.Payload)
}

func (o *DownloadParamsInternalServerError) GetPayload() *models.APIError {
	return o.Payload
}

func (o *DownloadParamsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
