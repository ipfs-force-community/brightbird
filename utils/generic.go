package utils

import (
	"encoding/json"

	"github.com/imdario/mergo"
)

func MergeStructAndJSON[T any](defaultCfg T, incomingCfg T, frontCfgBytes json.RawMessage) (T, error) {
	err := mergo.Merge(&defaultCfg, incomingCfg, mergo.WithOverride)
	if err != nil {
		return Default[T](), err
	}

	var frontCfg T
	err = json.Unmarshal(frontCfgBytes, &frontCfg)
	if err != nil {
		return Default[T](), err
	}
	err = mergo.Merge(&defaultCfg, frontCfg, mergo.WithOverride)
	if err != nil {
		return Default[T](), err
	}
	return defaultCfg, nil
}

func MergeStructAndInterface[T any](defaultCfg T, incomingCfg T, frontCfg interface{}) (T, error) {
	err := mergo.Merge(&defaultCfg, incomingCfg, mergo.WithOverride)
	if err != nil {
		return Default[T](), err
	}

	err = mergo.Merge(&defaultCfg, frontCfg, mergo.WithOverride)
	if err != nil {
		return Default[T](), err
	}
	return defaultCfg, nil
}

func Default[T any]() T {
	return *new(T)
}
