package api

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin/binding"
)

var paginationQueryBind = paginationQueryBindT{}

type paginationQueryBindT struct{}

func (paginationQueryBindT) Name() string {
	return "pagination_query"
}

func (paginationQueryBindT) Bind(req *http.Request, obj any) error {
	return binding.Query.Bind(req, obj)
	/*
		err := binding.Query.Bind(req, obj)
		if err != nil {
			return err
		}

		t := reflect.TypeOf(obj)
		if checkTypeIsPageReq(t.Elem()) { //todo reflect not support assert generic type
			paramsVal := reflect.ValueOf(obj).Elem().FieldByName("Params").Interface()
			return binding.Query.Bind(req, paramsVal)
		}
		return nil*/
}

// todo reflect not support assert generic type,
func checkTypeIsPageReq(t reflect.Type) bool {
	fieldNum := t.NumField()
	var check uint8
	for i := 0; i < fieldNum; i++ {
		field := t.Field(i)
		if field.Name == "PageNum" && field.Type.Kind() == reflect.Int64 {
			check = check | 0b1
		}
		if field.Name == "PageSize" && field.Type.Kind() == reflect.Int64 {
			check = check | 0b10
		}
		if field.Name == "Params" && field.Type.Kind() == reflect.Struct {
			check = check | 0b100
		}
	}
	return check == 0b111
}
