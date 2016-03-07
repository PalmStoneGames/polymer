package polymer

import (
	"reflect"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

func init() {
	js.Global.Set("_polymerGo_wrap", func(js *js.Object) *js.Object { return js })
}

type Encoder interface {
	// Encode returns the encoded object as a *js.Object and a boolean that is true if the object was non-empty
	Encode() (*js.Object, bool)
}

// Encode accepts a Go value and returns the encoded *js.Object, as well as a boolean that is true if the Go object was non-empty
func Encode(target interface{}) (*js.Object, bool) {
	return encodeRaw(reflect.ValueOf(target))
}

// encodeRaw accepts a reflect.Value and returns the encoded *js.Object, as well as a boolean that is true if the reflect.Value was non-empty
func encodeRaw(refVal reflect.Value) (*js.Object, bool) {
	// Special case for encoders
	if encoder, ok := refVal.Interface().(Encoder); ok {
		return encoder.Encode()
	}

	// Drill into interfaces
	if refVal.Kind() == reflect.Interface {
		refVal = refVal.Elem()
	}

	// Dril into pointers
	if refVal.Kind() == reflect.Ptr {
		refVal = refVal.Elem()
	}

	// Check validity
	if !refVal.IsValid() {
		return nil, false
	}

	refType := refVal.Type()
	switch refVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return InterfaceToJsObject(refVal.Interface()), reflect.Zero(refType).Interface() != refVal.Interface()
	case reflect.Slice:
		if refVal.Len() == 0 {
			return nil, false
		}

		s := js.S{}
		for i := 0; i < refVal.Len(); i++ {
			jsObj, _ := encodeRaw(refVal.Index(i))
			s = append(s, jsObj)
		}

		return InterfaceToJsObject(s), true
	case reflect.Struct:
		switch refVal.Interface().(type) {
		case time.Time:
			t := refVal.Interface().(time.Time)
			return InterfaceToJsObject(t), !t.IsZero()
		default:
			m := js.M{}
			return InterfaceToJsObject(m), encodeStruct(refVal, m)
		}

	default:
		return nil, false
	}
}

func encodeStruct(refVal reflect.Value, m js.M) bool {
	filled := false
	refType := refVal.Type()

	for i := 0; i < refType.NumField(); i++ {
		fieldType := refType.Field(i)
		if !isFieldExported(fieldType.Name) {
			continue
		}

		if fieldType.Anonymous && fieldType.Type != typeOfPtrBindProto {
			if (fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct) || fieldType.Type.Kind() == reflect.Struct {
				var localFilled bool
				if fieldType.Type.Kind() == reflect.Ptr {
					f := refVal.Field(i)
					if !f.IsNil() {
						localFilled = encodeStruct(f.Elem(), m)
					}
				} else {
					localFilled = encodeStruct(refVal.Field(i), m)
				}

				if localFilled {
					filled = true
				}

				continue
			}
		}

		jsObj, currFilled := encodeRaw(refVal.Field(i))
		if currFilled {
			filled = true
		}

		m[getJsName(fieldType.Name)] = jsObj
	}

	return filled
}

func InterfaceToJsObject(target interface{}) *js.Object {
	return js.Global.Call("_polymerGo_wrap", target)
}
