package types

import (
	"container/list"
	"reflect"
)

var _annotateOutType = reflect.TypeOf(AnnotateOut{})

type AnnotateOut struct {
}

func IsAnnotateOut(o interface{}) bool {
	return embedsType(o, _annotateOutType)
}

// Returns true if t embeds e or if any of the types embedded by t embed e.
func embedsType(i interface{}, e reflect.Type) bool {
	// TODO: this function doesn't consider e being a pointer.
	// given `type A foo { *In }`, this function would return false for
	// embedding dig.In, which makes for some extra error checking in places
	// that call this function. Might be worthwhile to consider reflect.Indirect
	// usage to clean up the callers.

	if i == nil {
		return false
	}

	// maybe it's already a reflect.Type
	t, ok := i.(reflect.Type)
	if !ok {
		// take the type if it's not
		t = reflect.TypeOf(i)
	}

	// We are going to do a breadth-first search of all embedded fields.
	types := list.New()
	types.PushBack(t)
	for types.Len() > 0 {
		t := types.Remove(types.Front()).(reflect.Type)

		if t == e {
			return true
		}

		if t.Kind() != reflect.Struct {
			continue
		}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				types.PushBack(f.Type)
			}
		}
	}

	// If perf is an issue, we can cache known In objects and Out objects in a
	// map[reflect.Type]struct{}.
	return false
}
