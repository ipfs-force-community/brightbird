package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
)

func ParserProperties(pathPrefix string, params reflect.Type) ([]types.Property, error) {
	if params.Kind() == reflect.Ptr {
		params = params.Elem()
	}

	numFields := params.NumField()
	properties := []types.Property{}
	for i := 0; i < numFields; i++ {
		field := params.Field(i)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if field.Anonymous {
			embedProperties, err := ParserProperties(pathPrefix, fieldType)
			if err != nil {
				return nil, err
			}
			properties = append(properties, embedProperties...)
			continue
		}

		if ignoreProperty(field) {
			continue
		}

		fieldName := getFieldJSONName(field)
		if fieldName == "-" || fieldName == "" {
			continue
		}

		description := field.Tag.Get("description")

		typeName, err := mapField(field)
		if err != nil {
			return nil, fmt.Errorf("field %s has unspport type %w", fieldName, err)
		}

		switch typeName {
		case "arrary":
			elemT := fieldType.Elem()
			fmt.Println(elemT.String())
			arrT, err := mapType(elemT)
			if err != nil {
				return nil, err
			}

			if arrT == "arrary" || arrT == "object" {
				//todo support complex type in arrary
				return nil, fmt.Errorf("arrary not support object or arrary")
			}

			fieldPath := joinPath(pathPrefix, fieldName)
			properties = append(properties, types.Property{
				Name:        fieldName,
				NamePath:    fieldPath,
				Type:        "arrary",
				Description: description,
				Chindren: []types.Property{
					{
						Name:        "[]", //todo index value in arrary
						NamePath:    fieldPath + "[]",
						Type:        arrT,
						Description: description,
					},
				},
			})
			continue
		case "object":
			if fieldType.Kind() == reflect.Struct {
				//json
				childProperties, err := ParserProperties(joinPath(pathPrefix, fieldName), fieldType)
				if err != nil {
					return nil, err
				}
				properties = append(properties, types.Property{
					Name:        fieldName,
					NamePath:    joinPath(pathPrefix, fieldName),
					Type:        "object",
					Description: description,
					Chindren:    childProperties,
				})
				continue
			}
			return nil, errors.New("wrong error definition")
		default:
			properties = append(properties, types.Property{
				Name:        fieldName,
				NamePath:    joinPath(pathPrefix, fieldName),
				Type:        typeName,
				Description: description,
			})
		}
	}
	return properties, nil
}

func joinPath(path, next string) string {
	if len(path) == 0 {
		return next
	}
	return path + "." + next
}

func ignoreProperty(field reflect.StructField) bool {
	return len(field.Tag.Get("ignore")) > 0
}

func getFieldJSONName(field reflect.StructField) string {
	fieldName := field.Name
	jsonTag := field.Tag.Get("json")
	jsonFlags := strings.Split(jsonTag, ",")
	if val := strings.TrimSpace(jsonFlags[0]); len(val) > 0 {
		fieldName = val
	}
	return fieldName
}

func mapField(val reflect.StructField) (string, error) {
	jsonType := val.Tag.Get("type")
	if len(jsonType) > 0 {
		return jsonType, nil
	}

	fieldType := val.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = val.Type.Elem()
	}

	return mapType(fieldType)
}

func mapType(t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.Bool:
		return "bool", nil
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Int64: //todo use bignumber
		fallthrough
	case reflect.Uint64: //todo use bignumber
		return "number", nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64: //todo use bigdecimal
		return "decimical", nil
	case reflect.String:
		return "string", nil
	case reflect.Struct:
		return "object", nil
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return "arrary", nil
	}
	return "", fmt.Errorf("types %s not support %s", t.String(), t.Kind())
}
func GetPropertyValue(property *types.Property, value string) (interface{}, error) {
	switch property.Type {
	case "bool":
		return value == "true", nil
	case "number": //todo consider big number
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "decimical":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "string":
		return value, nil
	case "object":
		var jsonRaw interface{}
		err := json.Unmarshal([]byte(value), &jsonRaw)
		if err != nil {
			return nil, err
		}
		return jsonRaw, nil
	}
	return nil, fmt.Errorf("unsupport property type %s", property.Type)
}

func ConvertValue[T any](value string) (T, error) {
	dstValue := new(T)
	valR := reflect.ValueOf(dstValue).Elem()
	switch valR.Type().Kind() {
	case reflect.Int:
		val, err := strconv.Atoi(value)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(val))
	case reflect.Bool:
		if value == "true" {
			valR.Set(reflect.ValueOf(true))
		} else {
			valR.Set(reflect.ValueOf(false))
		}

	case reflect.Int8:
		val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(int8(val)))

	case reflect.Int16:
		val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(int16(val)))

	case reflect.Int32:
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(int32(val)))

	case reflect.Int64: //todo use bignumber
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(val))

	case reflect.Uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(uint8(val)))

	case reflect.Uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(uint16(val)))

	case reflect.Uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(uint32(val)))

	case reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(val))
	case reflect.Float32:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(float32(val)))
	case reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return *dstValue, err
		}
		valR.Set(reflect.ValueOf(val))
	case reflect.String:
		valR.Set(reflect.ValueOf(value))
	default:
		return utils.Default[T](), fmt.Errorf("types %s not support", valR.Kind().String())
	}
	return *dstValue, nil
}
