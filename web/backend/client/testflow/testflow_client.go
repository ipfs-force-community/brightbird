// Code generated by go-swagger; DO NOT EDIT.

package testflow

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new testflow API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for testflow API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	Changetestflow(params *ChangetestflowParams, opts ...ClientOption) (*ChangetestflowOK, error)

	CountTestFlowRequest(params *CountTestFlowRequestParams, opts ...ClientOption) (*CountTestFlowRequestOK, error)

	DeleteTestFlow(params *DeleteTestFlowParams, opts ...ClientOption) (*DeleteTestFlowOK, error)

	GetTestFlowRequest(params *GetTestFlowRequestParams, opts ...ClientOption) (*GetTestFlowRequestOK, error)

	ListInGroupRequest(params *ListInGroupRequestParams, opts ...ClientOption) (*ListInGroupRequestOK, error)

	SaveTestFlow(params *SaveTestFlowParams, opts ...ClientOption) (*SaveTestFlowOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
Changetestflow change testflow group id
*/
func (a *Client) Changetestflow(params *ChangetestflowParams, opts ...ClientOption) (*ChangetestflowOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewChangetestflowParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "changetestflow",
		Method:             "POST",
		PathPattern:        "/changegroup",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ChangetestflowReader{formats: a.formats},
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
	success, ok := result.(*ChangetestflowOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for changetestflow: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CountTestFlowRequest Count testflow numbers in group
*/
func (a *Client) CountTestFlowRequest(params *CountTestFlowRequestParams, opts ...ClientOption) (*CountTestFlowRequestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCountTestFlowRequestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "countTestFlowRequest",
		Method:             "GET",
		PathPattern:        "/testflow/count",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CountTestFlowRequestReader{formats: a.formats},
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
	success, ok := result.(*CountTestFlowRequestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for countTestFlowRequest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
DeleteTestFlow Delete test flow by id
*/
func (a *Client) DeleteTestFlow(params *DeleteTestFlowParams, opts ...ClientOption) (*DeleteTestFlowOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteTestFlowParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deleteTestFlow",
		Method:             "DELETE",
		PathPattern:        "/testflow/{id}",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteTestFlowReader{formats: a.formats},
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
	success, ok := result.(*DeleteTestFlowOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for deleteTestFlow: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetTestFlowRequest gets specific test case by condition
*/
func (a *Client) GetTestFlowRequest(params *GetTestFlowRequestParams, opts ...ClientOption) (*GetTestFlowRequestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetTestFlowRequestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getTestFlowRequest",
		Method:             "GET",
		PathPattern:        "/testflow",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetTestFlowRequestReader{formats: a.formats},
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
	success, ok := result.(*GetTestFlowRequestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getTestFlowRequest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ListInGroupRequest lists test flows
*/
func (a *Client) ListInGroupRequest(params *ListInGroupRequestParams, opts ...ClientOption) (*ListInGroupRequestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListInGroupRequestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listInGroupRequest",
		Method:             "GET",
		PathPattern:        "/testflow/list",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListInGroupRequestReader{formats: a.formats},
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
	success, ok := result.(*ListInGroupRequestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listInGroupRequest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
SaveTestFlow save test case, create if not exist
*/
func (a *Client) SaveTestFlow(params *SaveTestFlowParams, opts ...ClientOption) (*SaveTestFlowOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSaveTestFlowParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "saveTestFlow",
		Method:             "POST",
		PathPattern:        "/testflow",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SaveTestFlowReader{formats: a.formats},
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
	success, ok := result.(*SaveTestFlowOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for saveTestFlow: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}