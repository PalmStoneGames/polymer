/*
Copyright 2015 Palm Stone Games, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package polymer

import (
	"reflect"
	"strings"
	"time"

	"fmt"
	"github.com/gopherjs/gopherjs/js"
)

// Decode decodes a js object to the target
// it watches for fields on the structure tagged with polymer-decode
// Tags can be of the following format: `polymer-decode:"js_field_name"`
func Decode(jsVal *js.Object, target interface{}) error {
	refType := reflect.TypeOf(target)
	if refType.Kind() != reflect.Ptr {
		return fmt.Errorf("target should be a pointer")
	}

	refType = refType.Elem()
	refVal := reflect.ValueOf(target).Elem()

	switch refVal.Kind() {
	case reflect.Int:
		refVal.Set(reflect.ValueOf(jsVal.Int()))
	case reflect.Int8:
		refVal.Set(reflect.ValueOf(int8(jsVal.Int())))
	case reflect.Int16:
		refVal.Set(reflect.ValueOf(int16(jsVal.Int())))
	case reflect.Int32:
		refVal.Set(reflect.ValueOf(int32(jsVal.Int())))
	case reflect.Int64:
		refVal.Set(reflect.ValueOf(jsVal.Int64()))
	case reflect.Uint:
		refVal.Set(reflect.ValueOf(uint(jsVal.Uint64())))
	case reflect.Uint8:
		refVal.Set(reflect.ValueOf(uint8(jsVal.Uint64())))
	case reflect.Uint16:
		refVal.Set(reflect.ValueOf(uint16(jsVal.Uint64())))
	case reflect.Uint32:
		refVal.Set(reflect.ValueOf(uint32(jsVal.Uint64())))
	case reflect.Uint64:
		refVal.Set(reflect.ValueOf(jsVal.Uint64()))
	case reflect.Float32:
		refVal.Set(reflect.ValueOf(float32(jsVal.Float())))
	case reflect.Float64:
		refVal.Set(reflect.ValueOf(jsVal.Float()))
	case reflect.String:
		refVal.Set(reflect.ValueOf(jsVal.String()))
	case reflect.Bool:
		refVal.Set(reflect.ValueOf(jsVal.Bool()))
	case reflect.Struct:
		switch refVal.Interface().(type) {
		case time.Time:
			timeMs := jsVal.Int64()
			refVal.Set(reflect.ValueOf(time.Unix(timeMs/1000, (timeMs%1000)*1000000)))
		default:
			return decodeStruct(jsVal, target)
		}
	default:
		return fmt.Errorf("Do not know how to deal with kind %v while decoding data for field %v", refVal.Kind(), refType.Name)
	}

	return nil
}

func decodeStruct(jsVal *js.Object, target interface{}) error {
	refVal := reflect.ValueOf(target)
	refType := reflect.TypeOf(target)
	if refType.Kind() != reflect.Ptr || refType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target should be a pointer to a struct")
	}

	refVal = refVal.Elem()
	refType = refType.Elem()

	for i := 0; i < refType.NumField(); i++ {
		// Grab field tag information
		fieldVal := refVal.Field(i)
		fieldType := refType.Field(i)

		// Check if the field is anonymous, if so, go through it as if it was at this level
		if fieldType.Anonymous {
			decodeStruct(jsVal, fieldVal.Addr().Interface())
			continue
		}

		tag := fieldType.Tag.Get("polymer-decode")
		if tag == "" {
			continue
		}

		// Get the actual value
		curr := jsVal
		for _, component := range strings.Split(tag, ".") {
			curr = curr.Get(component)
		}

		// Set the value
		if err := Decode(curr, fieldVal.Addr().Interface()); err != nil {
			return err
		}
	}

	return nil
}
