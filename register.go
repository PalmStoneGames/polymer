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

	"github.com/gopherjs/gopherjs/js"
)

const protoIndexKey = "_polymer_protoIndex"

//TODO: Use an opaque object set on this instead of a slice, the slice doesn't allow the proto nor js object to ever get freed
var jsMap []Interface

var (
	webComponentsReady   = false
	pendingRegistrations []js.M
)

func init() {
	js.Global.Get("window").Call("addEventListener", "WebComponentsReady", webComponentsReadyCallback)
}

// Register makes polymer aware of a certain type
// Polymer will analyze the type and use it for the tag returned by TagName()
// The type will then be instantiated automatically when tags corresponding to TagName are created through any method
func Register(proto Interface) {
	// Type detection
	refType := reflect.TypeOf(proto)
	if refType.Kind() != reflect.Ptr {
		panic("Expected proto to be a pointer to a struct")
	}
	if refType.Elem().Kind() != reflect.Struct {
		panic("Expected proto to be a pointer to a struct")
	}

	// Parse info
	handlers := parseHandlers(refType)
	tags := parseTags(refType.Elem())

	// Setup basics
	m := js.M{}
	m["is"] = proto.TagName()
	m["extends"] = proto.Extends()
	m["created"] = createdCallback(refType, tags)
	m["ready"] = readyCallback()
	m["attached"] = attachedCallback()
	m["detached"] = detachedCallback()

	// Setup properties
	properties := js.M{}
	for _, tag := range tags {
		if tag.Bind {
			curr := js.M{}
			curr["type"] = getJsType(refType.Elem().Field(tag.FieldIndex).Type)
			curr["notify"] = true

			properties[getJsName(tag.FieldName)] = curr
		}
	}
	m["properties"] = properties

	// Setup handlers
	for _, handler := range handlers {
		m[getJsName(handler.Name)] = handlerCallback(handler)
	}

	// Register our prototype with polymer
	if webComponentsReady {
		js.Global.Call("Polymer", m)
	} else {
		pendingRegistrations = append(pendingRegistrations, m)
	}
}

func webComponentsReadyCallback() {
	if !webComponentsReady {
		webComponentsReady = true
		for _, reg := range pendingRegistrations {
			js.Global.Call("Polymer", reg)
		}
	}
}
