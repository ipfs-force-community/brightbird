// Code generated by go-swagger; DO NOT EDIT.

package group

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new group API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for group API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CountGroup(params *CountGroupParams, opts ...ClientOption) (*CountGroupOK, error)

	DeleteGroup(params *DeleteGroupParams, opts ...ClientOption) (*DeleteGroupOK, error)

	GetGroupByID(params *GetGroupByIDParams, opts ...ClientOption) (*GetGroupByIDOK, error)

	ListGroup(params *ListGroupParams, opts ...ClientOption) (*ListGroupOK, error)

	SaveCases(params *SaveCasesParams, opts ...ClientOption) (*SaveCasesOK, error)

	UpdateGroup(params *UpdateGroupParams, opts ...ClientOption) (*UpdateGroupOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
CountGroup counts group by condition
*/
func (a *Client) CountGroup(params *CountGroupParams, opts ...ClientOption) (*CountGroupOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCountGroupParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "countGroup",
		Method:             "GET",
		PathPattern:        "/group/count",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CountGroupReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CountGroupOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for countGroup: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
DeleteGroup Delete group by id
*/
func (a *Client) DeleteGroup(params *DeleteGroupParams, opts ...ClientOption) (*DeleteGroupOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteGroupParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deleteGroup",
		Method:             "DELETE",
		PathPattern:        "/group/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteGroupReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteGroupOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for deleteGroup: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetGroupByID gets specific group by id
*/
func (a *Client) GetGroupByID(params *GetGroupByIDParams, opts ...ClientOption) (*GetGroupByIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetGroupByIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getGroupById",
		Method:             "GET",
		PathPattern:        "/group/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetGroupByIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetGroupByIDOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getGroupById: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ListGroup lists all group
*/
func (a *Client) ListGroup(params *ListGroupParams, opts ...ClientOption) (*ListGroupOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListGroupParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listGroup",
		Method:             "GET",
		PathPattern:        "/group/list",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListGroupReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListGroupOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listGroup: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
SaveCases Save group
*/
func (a *Client) SaveCases(params *SaveCasesParams, opts ...ClientOption) (*SaveCasesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSaveCasesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "saveCases",
		Method:             "POST",
		PathPattern:        "/group",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SaveCasesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*SaveCasesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for saveCases: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
UpdateGroup Update group name/show/description
*/
func (a *Client) UpdateGroup(params *UpdateGroupParams, opts ...ClientOption) (*UpdateGroupOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateGroupParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "updateGroup",
		Method:             "POST",
		PathPattern:        "/group/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UpdateGroupReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateGroupOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for updateGroup: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
