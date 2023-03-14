package types

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"sync"
)

type Category string

const (
	Deploy   Category = "Deployer"
	TestExec Category = "Exec"
)

type IPluginInfo interface {
	Each(fn func(*PluginDetail) error) error
	GetPlugin(name string) (*PluginDetail, error)
	AddPlugin(path string) error
}

type PluginInfo struct {
	Name        string
	Version     string
	Category    Category
	Description string
	Path        string
}

type PluginDetail struct {
	*PluginInfo
	Fn    reflect.Value
	Param reflect.Type
}

var _ IPluginInfo = (*PluginStore)(nil)

type PluginStore struct {
	lk      sync.Mutex
	plugins map[string]*PluginDetail
}

func NewExecPluginStore() *PluginStore {
	return &PluginStore{
		lk:      sync.Mutex{},
		plugins: make(map[string]*PluginDetail),
	}
}

func (store *PluginStore) Each(fn func(*PluginDetail) error) error {
	store.lk.Lock()
	defer store.lk.Unlock()
	for _, val := range store.plugins {
		err := fn(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *PluginStore) GetPlugin(name string) (*PluginDetail, error) {
	store.lk.Lock()
	defer store.lk.Unlock()
	plugin, ok := store.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

func (store *PluginStore) AddPlugin(path string) error {
	store.lk.Lock()
	defer store.lk.Unlock()
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	infoSymbol, err := p.Lookup("Info")
	if err != nil {
		return err
	}
	fnSymbol, err := p.Lookup("Exec")
	if err != nil {
		return err
	}

	rFn := reflect.ValueOf(fnSymbol)
	rfnT := reflect.TypeOf(fnSymbol)

	//check params and return
	if rfnT.NumIn() != 2 {
		return fmt.Errorf("plugin(%s) Exec must have 2 arguments", path)
	}

	if rfnT.In(0) != CtxT {
		return fmt.Errorf("plugin(%s) first argment must be context", path)
	}

	if rfnT.NumOut() == 1 {
		if rfnT.Out(0) != ErrT {
			return fmt.Errorf("plugin(%s) must have a error return ", path)
		}
	} else if rfnT.NumOut() == 2 {
		if rfnT.Out(1) != ErrT {
			return fmt.Errorf("plugin(%s) must have a error return ", path)
		}
	} else {
		return fmt.Errorf("plugin(%s) only have one or two return values and the last one must be error", path)
	}

	detail := &PluginDetail{
		PluginInfo: infoSymbol.(*PluginInfo),
		Fn:         rFn,
		Param:      rfnT.In(1),
	}

	store.plugins[detail.Name] = detail
	return nil
}

var ErrT = reflect.TypeOf((*error)(nil)).Elem()
var CtxT = reflect.TypeOf((*context.Context)(nil)).Elem()
var NilVal = reflect.ValueOf(nil)

func LoadPlugins(dir string) (*PluginStore, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pluginStore := NewExecPluginStore()
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".so") {
				err = pluginStore.AddPlugin(filepath.Join(dir, entry.Name()))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return pluginStore, nil
}
