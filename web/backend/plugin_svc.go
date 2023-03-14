package main

import "github.com/hunjixin/brightbird/types"

type IPluginService interface {
	List(interface{}, error)
}

type PluginSvc struct {
	deployPluginStore types.PluginStore
	execPluginStore   types.PluginStore
}

type PluginDetail struct {
	Name     string
	Category string
}

type ListOutput struct {
	Deploy types.PluginDetail
	Cases  types.PluginDetail
}

func (p *PluginSvc) List(i interface{}, err error) {

}

func NewPluginSvc() IPluginService {
	return &PluginSvc{}
}
