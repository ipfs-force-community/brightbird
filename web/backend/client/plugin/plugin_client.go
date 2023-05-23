// Code generated by go-swagger; DO NOT EDIT.

package plugin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new plugin API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for plugin API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	DeletePlugin(params *DeletePluginParams, opts ...ClientOption) (*DeletePluginOK, error)

	GetPlugin(params *GetPluginParams, opts ...ClientOption) (*GetPluginOK, error)

	GetPluginMainfest(params *GetPluginMainfestParams, opts ...ClientOption) (*GetPluginMainfestOK, error)

	ImportPlugin(params *ImportPluginParams, opts ...ClientOption) (*ImportPluginOK, error)

	UploadPluginFilesParams(params *UploadPluginFilesParamsParams, opts ...ClientOption) (*UploadPluginFilesParamsOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
DeletePlugin Delete plugin by id
*/
func (a *Client) DeletePlugin(params *DeletePluginParams, opts ...ClientOption) (*DeletePluginOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeletePluginParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deletePlugin",
		Method:             "DELETE",
		PathPattern:        "/plugin",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeletePluginReader{formats: a.formats},
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
	success, ok := result.(*DeletePluginOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for deletePlugin: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetPlugin gets plugin by name and version
*/
func (a *Client) GetPlugin(params *GetPluginParams, opts ...ClientOption) (*GetPluginOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPluginParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPlugin",
		Method:             "GET",
		PathPattern:        "/plugin",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetPluginReader{formats: a.formats},
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
	success, ok := result.(*GetPluginOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getPlugin: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetPluginMainfest gets plugin mainfest
*/
func (a *Client) GetPluginMainfest(params *GetPluginMainfestParams, opts ...ClientOption) (*GetPluginMainfestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPluginMainfestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPluginMainfest",
		Method:             "GET",
		PathPattern:        "/plugin/mainfest",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetPluginMainfestReader{formats: a.formats},
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
	success, ok := result.(*GetPluginMainfestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getPluginMainfest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ImportPlugin imports plugin mainfest
*/
func (a *Client) ImportPlugin(params *ImportPluginParams, opts ...ClientOption) (*ImportPluginOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewImportPluginParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "importPlugin",
		Method:             "POST",
		PathPattern:        "/plugin/import",
		ProducesMediaTypes: []string{"application/json", "application/text"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ImportPluginReader{formats: a.formats},
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
	success, ok := result.(*ImportPluginOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for importPlugin: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
UploadPluginFilesParams Upload plugin files
*/
func (a *Client) UploadPluginFilesParams(params *UploadPluginFilesParamsParams, opts ...ClientOption) (*UploadPluginFilesParamsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUploadPluginFilesParamsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "uploadPluginFilesParams",
		Method:             "POST",
		PathPattern:        "/plugin/upload",
		ProducesMediaTypes: []string{"application/json", "application/xml"},
		ConsumesMediaTypes: []string{"application/json", "application/xml"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UploadPluginFilesParamsReader{formats: a.formats},
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
	success, ok := result.(*UploadPluginFilesParamsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for uploadPluginFilesParams: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
