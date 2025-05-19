package log

import "fmt"

//go:generate stringer -type=PluginType -linecomment
type PluginType int

const (
	// InputPlugin input
	InputPlugin PluginType = iota
	// FilterPlugin filter
	FilterPlugin
	// OutputPlugin output
	OutputPlugin
)

var _ fmt.Stringer = PluginType(0)
