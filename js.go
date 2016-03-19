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
	"time"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
)

var propertyEventNameRegExp *regexp.Regexp

func init() {
	var err error
	propertyEventNameRegExp, err = regexp.Compile("([a-z])([A-Z])")
	if err != nil {
		panic(err)
	}
}

const DateTimeLocalFormat = "2006-01-02T15:04:05"

// DateTimeLocal is a time.Time that properly encodes and decodes to the format expected datetime-local input fields
// The format is as follows, as per the format definition of the time package: "2006-01-02T15:04:05"
type DateTimeLocal time.Time

type AsyncHandle struct {
	jsHandle *js.Object
}

func (t DateTimeLocal) Encode() (*js.Object, bool) {
	return InterfaceToJsObject(time.Time(t).Local().Format(DateTimeLocalFormat)), !time.Time(t).IsZero()
}

func (t *DateTimeLocal) Decode(val *js.Object) error {
	if val == js.Undefined || val == nil || val.String() == "" {
		*t = DateTimeLocal{}
		return nil
	}

	parsedTime, err := time.ParseInLocation(DateTimeLocalFormat, val.String(), time.Local)
	if err != nil {
		return err
	}

	*t = DateTimeLocal(parsedTime)
	return nil
}

func Log(args ...interface{}) {
	js.Global.Get("console").Call("log", args...)
}

func getJsName(fieldName string) string {
	endIndex := len(fieldName) - 1
	newFieldName := ""
	for i, rune := range fieldName {
		newFieldName += string(unicode.ToLower(rune))
		if unicode.IsLower(rune) {
			endIndex = i
			break
		}
	}

	newFieldName += fieldName[endIndex+1:]
	return newFieldName
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
	return js.Global.Get("Polymer").Get("CaseMap").Call("camelToDashCase", fieldName).String() + "-changed"
}

// Async calls the given callback asynchronously.
// If the specified wait time is -1, the callback will be ran with microtask timing (after the current method finishes, but before the next event from the event queue is processed)
// Otherwise, its ran waitTime milliseconds in the future. A waitTime of 1 can be useful to run a callback after all events currently in the queue have been processed.
// Returns a handle that can be used to cancel the task
func Async(waitTime int, f func()) *AsyncHandle {
	handle := &AsyncHandle{}
	if waitTime == -1 {
		handle.jsHandle = js.Global.Get("Polymer").Get("Async").Call("run", f)
	} else {
		handle.jsHandle = js.Global.Get("Polymer").Get("Async").Call("run", f, waitTime)
	}

	return handle
}

func (p *Proto) CancelAsync(handle *AsyncHandle) {
	js.Global.Get("Polymer").Get("Async").Call("cancel", handle.jsHandle)
}
