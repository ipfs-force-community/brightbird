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

// ListDeployPluginsReader is a Reader for the ListDeployPlugins structure.
type ListDeployPluginsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListDeployPluginsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListDeployPluginsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 503:
		result := NewListDeployPluginsServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListDeployPluginsOK creates a ListDeployPluginsOK with default headers values
func NewListDeployPluginsOK() *ListDeployPluginsOK {
	return &ListDeployPluginsOK{}
}

/*
ListDeployPluginsOK describes a response with status code 200, with default header values.

pluginDetail
*/
type ListDeployPluginsOK struct {
	Payload []*models.PluginDetail
}

// IsSuccess returns true when this list deploy plugins o k response has a 2xx status code
func (o *ListDeployPluginsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list deploy plugins o k response has a 3xx status code
func (o *ListDeployPluginsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list deploy plugins o k response has a 4xx status code
func (o *ListDeployPluginsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list deploy plugins o k response has a 5xx status code
func (o *ListDeployPluginsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list deploy plugins o k response a status code equal to that given
func (o *ListDeployPluginsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list deploy plugins o k response
func (o *ListDeployPluginsOK) Code() int {
	return 200
}

func (o *ListDeployPluginsOK) Error() string {
	return fmt.Sprintf("[GET /plugin/deploy][%d] listDeployPluginsOK  %+v", 200, o.Payload)
}

func (o *ListDeployPluginsOK) String() string {
	return fmt.Sprintf("[GET /plugin/deploy][%d] listDeployPluginsOK  %+v", 200, o.Payload)
}

func (o *ListDeployPluginsOK) GetPayload() []*models.PluginDetail {
	return o.Payload
}

func (o *ListDeployPluginsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListDeployPluginsServiceUnavailable creates a ListDeployPluginsServiceUnavailable with default headers values
func NewListDeployPluginsServiceUnavailable() *ListDeployPluginsServiceUnavailable {
	return &ListDeployPluginsServiceUnavailable{}
}

/*
ListDeployPluginsServiceUnavailable describes a response with status code 503, with default header values.

apiError
*/
type ListDeployPluginsServiceUnavailable struct {
	Payload *models.APIError
}

// IsSuccess returns true when this list deploy plugins service unavailable response has a 2xx status code
func (o *ListDeployPluginsServiceUnavailable) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list deploy plugins service unavailable response has a 3xx status code
func (o *ListDeployPluginsServiceUnavailable) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list deploy plugins service unavailable response has a 4xx status code
func (o *ListDeployPluginsServiceUnavailable) IsClientError() bool {
	return false
}

// IsServerError returns true when this list deploy plugins service unavailable response has a 5xx status code
func (o *ListDeployPluginsServiceUnavailable) IsServerError() bool {
	return true
}

// IsCode returns true when this list deploy plugins service unavailable response a status code equal to that given
func (o *ListDeployPluginsServiceUnavailable) IsCode(code int) bool {
	return code == 503
}

// Code gets the status code for the list deploy plugins service unavailable response
func (o *ListDeployPluginsServiceUnavailable) Code() int {
	return 503
}

func (o *ListDeployPluginsServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /plugin/deploy][%d] listDeployPluginsServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListDeployPluginsServiceUnavailable) String() string {
	return fmt.Sprintf("[GET /plugin/deploy][%d] listDeployPluginsServiceUnavailable  %+v", 503, o.Payload)
}

func (o *ListDeployPluginsServiceUnavailable) GetPayload() *models.APIError {
	return o.Payload
}

func (o *ListDeployPluginsServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
