package main

import (
	"errors"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/saagie/fluent-bit-mongo/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type Value struct {
	Logger log.Logger
	Config interface{}
	Db     *mongo.Database
}

func Get(ctxPointer unsafe.Pointer) (*Value, error) {
	value := output.FLBPluginGetContext(ctxPointer)
	if value == nil {
		return &Value{}, errors.New("no value found")
	}

	return value.(*Value), nil
}

func Set(ctxPointer unsafe.Pointer, value *Value) {
	output.FLBPluginSetContext(ctxPointer, value)
}
