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

	"github.com/gopherjs/gopherjs/js"
)

type fieldTag struct {
	FieldIndex int
	FieldName  string
}

// Register makes polymer aware of a certain type
// Polymer will analyze the type and use it for the tag returned by TagName()
// The type will then be instantiated automatically when tags corresponding to TagName are created through any method
func Register(proto Interface) {
	// Type detection
	refType := reflect.TypeOf(proto)
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}

	// Setup js.M object
	m := js.M{}
	m["is"] = proto.TagName()
	m["extends"] = proto.Extends()
	m["created"] = createdCallback(refType)

	// Register our prototype with polymer
	js.Global.Call("Polymer", m)
}

func parseTags(refType reflect.Type) []*fieldTag {
	var tags []*fieldTag
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		tag := strings.Split(field.Tag.Get("polymer"), ",")

		// First field in the tag is the name, if it isn't present, we bail
		if len(tag) == 0 || tag[0] == "" {
			continue
		}

		tags = append(tags, &fieldTag{
			FieldIndex: i,
			FieldName:  tag[0],
		})
	}

	return tags
}

func createdCallback(refType reflect.Type) *js.Object {
	tags := parseTags(refType)

	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// Create a new Go side object and keep it around in this closure
		// That way, we can keep track of it across callbacks and calls
		refVal := reflect.New(refType)
		proto := refVal.Interface().(Interface)
		refVal = refVal.Elem()

		// Trigger the Created callback
		proto.Created()

		// Setup Ready callback
		this.Set("ready", func() {
			for _, tag := range tags {
				this.Set(tag.FieldName, refVal.Field(tag.FieldIndex).Interface())
			}

			proto.Ready()
		})

		this.Set("attached", func() { proto.Attached() })
		this.Set("detached", func() { proto.Detached() })

		return nil
	})
}
