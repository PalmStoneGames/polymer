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
	"regexp"

	"github.com/gopherjs/gopherjs/js"
	"strings"
)

var propertyEventNameRegExp *regexp.Regexp

func init() {
	var err error
	propertyEventNameRegExp, err = regexp.Compile("([a-z])([A-Z])")
	if err != nil {
		panic(err)
	}
}

func getJsType(t reflect.Type) *js.Object {
	switch t.Kind() {
	case reflect.String:
		return js.Global.Get("String")
	case reflect.Bool:
		return js.Global.Get("Boolean")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return js.Global.Get("Number")
	default:
		return js.Global.Get("Object")
	}
}

func getJsPropertyChangedEvent(fieldName string) string {
	return strings.ToLower(js.Global.Get("Polymer").Get("CaseMap").Call("camelToDashCase", fieldName).String()) + "-changed"
}
