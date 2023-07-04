package main

import (
	"errors"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func IterJSON(iter *jsoniter.Iterator, encoder *jsoniter.Stream, fieldPath string, valResolve func(string, string) (interface{}, error)) error {
	switch iter.WhatIsNext() {
	case jsoniter.InvalidValue:
		return errors.New("invalidate json")
	case jsoniter.StringValue:
		val, err := valResolve(strings.Trim(fieldPath, "."), iter.ReadString()) //express or convert to by property type
		if err != nil {
			return err
		}
		encoder.WriteVal(val)
	case jsoniter.BoolValue:
		encoder.WriteVal(iter.ReadBool())
	case jsoniter.NilValue:
		encoder.WriteVal(iter.ReadNil())
	case jsoniter.NumberValue:
		number := iter.ReadNumber()
		encoder.WriteVal(number)
	case jsoniter.ObjectValue:
		encoder.WriteObjectStart()

		hasWrite := false
		iter.ReadObjectCB(func(objIter *jsoniter.Iterator, s string) bool {
			encoder.WriteObjectField(s)
			err := IterJSON(objIter, encoder, fieldPath+"."+s, valResolve)
			if err != nil {
				objIter.ReportError("iter", "resolve object fail")
				return false
			}
			encoder.WriteMore()
			hasWrite = true
			return true
		})
		if hasWrite {
			buf := encoder.Buffer()
			encoder.SetBuffer(buf[:len(buf)-1])
		}
		encoder.WriteObjectEnd()
	case jsoniter.ArrayValue:
		return errors.New("arrary not support")
	default:
		return errors.New("not support type")
	}
	return nil
}
