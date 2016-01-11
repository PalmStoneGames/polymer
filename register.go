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
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

const protoIndexKey = "_polymer_protoIndex"

var (
	//TODO: Use an opaque object set on this instead of a slice, the slice doesn't allow the proto nor js object to ever get freed
	jsMap                  []Interface
	webComponentsReady     = false
	pendingGoRegistrations = make(map[string]js.M)
	pendingJSRegistrations []string
	onReadyChans           []chan struct{}
)

// CustomRegistrationAttrs can be used to pass custom attributes to the generated javascript prototype
type CustomRegistrationAttr struct {
	Name  string
	Value interface{}
}

type ElementDefinition struct {
	protoDef js.M
}

func init() {
	// Setup a global that can be called from js to register an element with us
	js.Global.Set("PolymerGo", polymerGo)

	// Listen to the WebComponentsReady callback to actually register our events
	js.Global.Get("window").Call("addEventListener", "WebComponentsReady", webComponentsReadyCallback)
}

func lookupProto(obj *js.Object) Interface {
	index := obj.Get(protoIndexKey)
	if index == js.Undefined || index == nil {
		panic("protoIndexKey not found")
	}

	return jsMap[index.Int()]
}

// WithExtends can be passed as option to Register to make an element extend another element
func WithExtends(extends string) CustomRegistrationAttr {
	return CustomRegistrationAttr{
		Name:  "extends",
		Value: extends,
	}
}

// WithBehaviors can be used to pass the protos of other polymer objects to be
// - If a string is passed, polymer will look for a global with that name,
// - If a *ElementDefinition is passed, it will be used as is
// - If a *js.Object is passed, it is set directly as behavior
func WithBehaviors(behaviors ...interface{}) CustomRegistrationAttr {
	return CustomRegistrationAttr{
		Name:  "behaviors",
		Value: behaviors,
	}
}

// Register makes polymer aware of a certain type
// Polymer will analyze the type and use it for the tag returned by TagName()
// The type will then be instantiated automatically when tags corresponding to TagName are created through any method
func Register(tagName string, proto Interface, customAttrs ...CustomRegistrationAttr) *ElementDefinition {
	if webComponentsReady {
		panic("polymer.Register call after WebComponentsReady has triggered")
	}

	if !strings.Contains(tagName, "-") {
		panic("Tagnames must contain a dash according to polymer's standards for custom elements")
	}

	// Type detection
	refType := reflect.TypeOf(proto)
	if refType.Kind() != reflect.Ptr {
		panic("Expected proto to be a pointer to a struct")
	}
	if refType.Elem().Kind() != reflect.Struct {
		panic("Expected proto to be a pointer to a struct")
	}

	// Setup basics
	m := js.M{}
	m["is"] = tagName
	m["created"] = createdCallback(refType)
	m["ready"] = readyCallback()
	m["attached"] = attachedCallback()
	m["detached"] = detachedCallback()
	m["properties"] = parseProperties(refType)

	// Setup handlers
	for _, handler := range parseHandlers(refType) {
		m[getJsName(handler.Name)] = eventHandlerCallback(handler.Func)
	}

	// Setup compute functions
	for _, handler := range parseComputes(refType) {
		m[getJsName(handler.Name)] = computeCallback(handler.Func)
	}

	// Note: Channel based event handlers are not setup here, they're setup in Created() as we need to actually make the channels

	// Setup observers
	setObservers(refType, m)

	// Custom attributes
	for _, attr := range customAttrs {
		m[attr.Name] = attr.Value
	}

	// Register our prototype with polymer
	pendingGoRegistrations[tagName] = m
	return &ElementDefinition{m}
}

// OnReady returns a channel that will be closed once polymer has been initialized
func OnReady() <-chan struct{} {
	c := make(chan struct{})
	onReadyChans = append(onReadyChans, c)
	return c
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

	sort.Sort(sort.StringSlice(goTagNames))
	sort.Sort(sort.StringSlice(jsTagNames))

	i = 0
	j := 0
	var jsOnlyRegistration []string
	var goOnlyRegistration []string
	var fullRegistration []string
	for i < len(goTagNames) && j < len(jsTagNames) {
		if goTagNames[i] < jsTagNames[j] { // Present in go but not js
			goOnlyRegistration = append(goOnlyRegistration, goTagNames[i])
			i++
		} else if jsTagNames[j] < goTagNames[i] { // Present in js but not go
			jsOnlyRegistration = append(jsOnlyRegistration, jsTagNames[j])
			j++
		} else { // Present in both
			fullRegistration = append(fullRegistration, goTagNames[i])
			i++
			j++
		}
	}
	for ; i < len(goTagNames); i++ {
		goOnlyRegistration = append(goOnlyRegistration, goTagNames[i])
	}
	for ; j < len(jsTagNames); j++ {
		jsOnlyRegistration = append(jsOnlyRegistration, jsTagNames[j])
	}

	// Check for JS only registrations, those are invalid
	for _, tagName := range jsOnlyRegistration {
		fmt.Printf("Error: '%v' is registered through PolymerGo(), but polymer.Register() was never called for it.\n", tagName)
	}
	if len(jsOnlyRegistration) != 0 {
		panic("All tags registered through PolymerGo must have polymer.Register() called for them")
	}

	// process all JS registrations in order, and then all the go only ones
	// Loop through the JS registrations and call Polymer()
	for _, tagName := range pendingJSRegistrations {
		doRegister(pendingGoRegistrations[tagName])
	}

	// Register Go only elements as well
	for _, tagName := range goOnlyRegistration {
		doRegister(pendingGoRegistrations[tagName])
	}

	// Close all ready chans
	for _, c := range onReadyChans {
		close(c)
	}
}

func doRegister(protoDef js.M) {
	if protoDef["behaviors"] != nil {
		behaviors := protoDef["behaviors"].([]interface{})
		for i, val := range behaviors {
			switch val.(type) {
			case string:
				name := val.(string)
				global := js.Global.Get(name)
				if global != nil && global != js.Undefined {
					behaviors[i] = global
				} else {
					behaviors[i] = js.Global.Get("Polymer").Get(name)
				}
			case *ElementDefinition:
				behaviors[i] = val.(*ElementDefinition).protoDef
			case *js.Object:
			default:
				panic(fmt.Sprintf("Don't know what to do with behavior of type %T", behaviors[i]))
			}
		}

		protoDef["behaviors"] = behaviors
	}

	js.Global.Call("Polymer", protoDef)
}

func polymerGo(tagName string) {
	if webComponentsReady {
		panic("PolymerGo call after WebComponentsReady has triggered")
	}

	if !strings.Contains(tagName, "-") {
		panic("Tagnames must contain a dash according to polymer's standards for custom elements")
	}

	pendingJSRegistrations = append(pendingJSRegistrations, tagName)
}

func parseProperties(refType reflect.Type) js.M {
	properties := js.M{}

	refType = refType.Elem()
	for i := 0; i < refType.NumField(); i++ {
		fieldType := refType.Field(i)
		if fieldType.Anonymous && (fieldType.Type == typeOfPtrProto || fieldType.Type == typeOfPtrBindProto) {
			continue
		}

		tagText := fieldType.Tag.Get("polymer")
		if tagText == "" {
			continue
		}

		tag := strings.Split(tagText, ",")
		for i := 0; i < len(tag); i++ {
			switch tag[i] {
			case "bind":
				properties[getJsName(fieldType.Name)] = js.M{
					"type":   getJsType(refType.FieldByIndex(fieldType.Index).Type),
					"notify": true,
				}
			}
		}
	}

	return properties
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

func parseChanHandlers(refType reflect.Type) []reflect.StructField {
	refType = refType.Elem()
	var handlers []reflect.StructField

	for i := 0; i < refType.NumField(); i++ {
		fieldType := refType.Field(i)
		if fieldType.Anonymous && (fieldType.Type == typeOfPtrProto || fieldType.Type == typeOfPtrBindProto) {
			continue
		}

		tagText := fieldType.Tag.Get("polymer")
		if tagText == "" {
			continue
		}

		tag := strings.Split(tagText, ",")
		for i := 0; i < len(tag); i++ {
			switch tag[i] {
			case "handler":
				handlers = append(handlers, fieldType)
			}
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

func setObservers(refType reflect.Type, m js.M) {
	observers := js.S{}
	setObserversNested(refType.Elem(), m, &observers, nil)
	m["observers"] = observers
}

func setObserversNested(refType reflect.Type, m js.M, observers *js.S, path []string) {
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)

		// If we're dealing with *polymer.Proto, skip
		if field.Anonymous && field.Type == typeOfPtrProto {
			continue
		}

		currPath := make([]string, len(path)+1)
		copy(currPath, path)
		currPath[len(path)] = getJsName(field.Name)

		// Check if this field has the bind fieldtag
		bind := false
		for _, tag := range strings.Split(field.Tag.Get("polymer"), ",") {
			if tag == "bind" {
				bind = true
				break
			}
		}

		// Figure out the kind and type
		fieldType := field.Type
		// Skip *js.Object
		if fieldType == typeOfJsObject {
			continue
		}

		// Deserialize pointer
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Use the kind to decide what to do
		switch fieldType.Kind() {
		case reflect.Struct:
			if bind {
				funcName, bindStr := pathBind(currPath, "*")
				*observers = append(*observers, bindStr)
				m[funcName] = observeDeepCallback()
			}
			setObserversNested(fieldType, m, observers, currPath)
		case reflect.Slice:
			if bind {
				funcName, bindStr := pathBind(currPath, "*")
				*observers = append(*observers, bindStr)
				m[funcName] = observeDeepCallback()
			}
		default:
			// Add the current field if bound
			if bind {
				funcName, bindStr := pathBind(currPath, "")
				*observers = append(*observers, bindStr)
				m[funcName] = observeShallowCallback(currPath)
			}
		}
	}

	return
}

func pathBind(path []string, mode string) (string, string) {
	funcName := fmt.Sprintf("observe_%v", strings.Join(path, "_"))
	bindStr := strings.Join(path, ".")

	if mode != "" {
		bindStr = fmt.Sprintf("%v.%v", bindStr, mode)
	}

	return funcName, fmt.Sprintf("%v(%v)", funcName, bindStr)
}
