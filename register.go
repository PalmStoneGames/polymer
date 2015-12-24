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

	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/xtgo/set"
	"sort"
	"strings"
)

const protoIndexKey = "_polymer_protoIndex"

//TODO: Use an opaque object set on this instead of a slice, the slice doesn't allow the proto nor js object to ever get freed
var jsMap []Interface

var (
	webComponentsReady     = false
	pendingGoRegistrations = make(map[string]js.M)
	pendingJSRegistrations []string
)

func init() {
	// Setup a global that can be called from js to register an element with us
	js.Global.Set("PolymerGo", polymerGoCallback)

	// Listen to the WebComponentsReady callback to actually register our events
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

			properties[tag.FieldName] = curr
		}
	}
	m["properties"] = properties

	// Setup handlers
	for _, handler := range parseHandlers(refType) {
		m[handler.Name] = handlerCallback(handler)
	}

	// Setup compute functions
	for _, handler := range parseComputes(refType) {
		m[handler.Name] = computeCallback(handler)
	}

	// Register our prototype with polymer
	if webComponentsReady {
		panic("polymer.Register call after WebComponentsReady has triggered")
	} else {
		pendingGoRegistrations[proto.TagName()] = m
	}
}

func webComponentsReadyCallback() {
	webComponentsReady = true

	// Get all tag names for our Go side registrations
	goTagNames := make([]string, len(pendingGoRegistrations))
	i := 0
	for _, reg := range pendingGoRegistrations {
		goTagNames[i] = reg["is"].(string)
		i++
	}

	// Setup a copy of the pendingJSRegistrations slice for use with the set package
	// We don't want to lose our ordering
	jsTagNames := make([]string, len(pendingJSRegistrations))
	copy(jsTagNames, pendingJSRegistrations)

	// Set up a set for the go and for the JS tag names
	goSet := set.Strings(goTagNames)
	jsSet := set.Strings(jsTagNames)

	// Merge both and set up the pivots for use with the set package
	var data []string
	data = append(data, goSet...)
	data = append(data, jsSet...)

	diffs := data[:set.Diff(sort.StringSlice(data), len(goSet))]
	if len(diffs) != 0 {
		for _, diff := range diffs {
			fmt.Printf("%v was not registered correctly\n", diff)
		}
		panic("Expected all registrations to be complete by the time the WebComponentsReady event triggers")
	}

	// Loop through the JS registrations and call Polymer()
	for _, tagName := range pendingJSRegistrations {
		js.Global.Call("Polymer", pendingGoRegistrations[tagName])
	}
}

func polymerGoCallback(tagName string) {
	if webComponentsReady {
		panic("PolymerGo call after WebComponentsReady has triggered")
	}

	pendingJSRegistrations = append(pendingJSRegistrations, tagName)
}

type fieldTag struct {
	FieldIndex int
	FieldName  string
	Bind       bool
}

func parseTags(refType reflect.Type) []*fieldTag {
	var tags []*fieldTag
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		tagText := field.Tag.Get("polymer")
		if tagText == "" {
			continue
		}
		tag := strings.Split(tagText, ",")

		f := fieldTag{
			FieldIndex: i,
			FieldName:  field.Name,
		}

		for i := 0; i < len(tag); i++ {
			switch tag[i] {
			case "bind":
				f.Bind = true
			}
		}

		tags = append(tags, &f)
	}

	return tags
}

func parseHandlers(refType reflect.Type) []reflect.Method {
	var handlers []reflect.Method

	for i := 0; i < refType.NumMethod(); i++ {
		method := refType.Method(i)

		if strings.HasPrefix(method.Name, "Handle") {
			handlers = append(handlers, method)
		}
	}

	return handlers
}

func parseComputes(refType reflect.Type) []reflect.Method {
	var handlers []reflect.Method

	for i := 0; i < refType.NumMethod(); i++ {
		method := refType.Method(i)

		if strings.HasPrefix(method.Name, "Compute") {
			handlers = append(handlers, method)
		}
	}

	return handlers
}
