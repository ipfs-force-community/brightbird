package types

import "github.com/hunjixin/brightbird/env/types"

type PluginType = types.PluginType
type Endpoint = types.Endpoint
type PluginInfo = types.PluginInfo
type BootstrapPeers = types.BootstrapPeers

var EndpointFromString = types.EndpointFromString
var EndpointFromHostPort = types.EndpointFromHostPort

const OutLabel = types.OutLabel
const SvcName = types.SvcName
const CodeVersion = types.CodeVersion

var GetString = types.GetString
var PtrString = types.PtrString

type TestId string

type PrivateRegistry string

type Shutdown chan struct{}
