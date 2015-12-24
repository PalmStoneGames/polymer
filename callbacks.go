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

	"github.com/gopherjs/gopherjs/js"
)

func createdCallback(refType reflect.Type, tags []*fieldTag) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// Create a new Go side object
		refVal := reflect.New(refType.Elem())
		proto := refVal.Interface().(Interface)
		refVal = refVal.Elem()

		// Set the proto value, this is needed because we get our callers to embed *polymer.Proto, and it needs to get instantiated
		refVal.FieldByName("Proto").Set(reflect.ValueOf(&Proto{}))

		// Store ourselves in js land so we can map js to proto
		jsMap = append(jsMap, proto)
		this.Set(protoIndexKey, len(jsMap)-1)

		// Set data on the proto
		data := proto.data()
		data.this = this
		data.Element = WrapJSElement(this)
		data.tags = tags

		// Call the proto side callback for user hooks
		proto.Created()

		return nil
	})
}

func readyCallback() *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// Lookup the proto
		proto := jsMap[this.Get(protoIndexKey).Int()]
		refVal := reflect.ValueOf(proto)

		// Set initial field values and register change events
		for _, tag := range proto.data().tags {
			// Get field info first
			fieldVal := refVal.Elem().Field(tag.FieldIndex)
			fieldType := fieldVal.Type()

			// Set the value on a dummy first, we compare it with the zero value after to decide whether to set the field from js or read it from Go
			currVal := reflect.New(fieldType)
			zeroVal := reflect.Zero(fieldType)
			if err := Decode(this.Get(tag.FieldName), currVal.Interface()); err != nil {
				panic(fmt.Sprintf("Error while decoding polymer field value for %v: %v", tag.FieldName, err))
			}

			// Decide whether to do js -> go or go -> js
			// If the value read from js is the zero value, we do go -> js
			// otherwise, we do js -> go
			if currVal.Elem().Interface() == zeroVal.Interface() {
				this.Set(tag.FieldName, fieldVal.Interface())
			} else {
				fieldVal.Set(currVal.Elem())
			}

			if tag.Bind {
				this.Call("addEventListener", getJsPropertyChangedEvent(tag.FieldName), propertyChangeCallback(tag))
			}
		}

		// Call the proto side callback for user hooks
		proto.Ready()

		return nil
	})
}

func attachedCallback() *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// Lookup the proto
		proto := jsMap[this.Get(protoIndexKey).Int()]

		// Call the proto side callback for user hooks
		proto.Attached()

		return nil
	})
}

func detachedCallback() *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// Lookup the proto
		proto := jsMap[this.Get(protoIndexKey).Int()]

		// Call the proto side callback for user hooks
		proto.Detached()

		return nil
	})
}

func propertyChangeCallback(tag *fieldTag) *js.Object {
	return js.MakeFunc(func(this *js.Object, jsArgs []*js.Object) interface{} {
		// Fetch the proto and refVal
		proto := jsMap[this.Get(protoIndexKey).Int()]
		refVal := reflect.ValueOf(proto)

		// Decode the event
		var e PropertyChangedEvent
		if err := Decode(jsArgs[0], &e); err != nil {
			panic(fmt.Sprintf("Error while decoding event: %v", err))
		}

		// Decode the value, it's left undecoded in the event because it can't be typed
		// We have the field type however, so we can use the *js.Object for the value and then set the interface{} val
		fieldType := reflect.TypeOf(proto).Elem().Field(tag.FieldIndex).Type
		fieldVal := reflect.New(fieldType)
		if err := Decode(e.JSValue, fieldVal.Interface()); err != nil {
			panic(fmt.Sprintf("Error while decoding value: %v", err))
		}

		// Set the field on the Go side struct and on the event
		refVal.Elem().Field(tag.FieldIndex).Set(fieldVal.Elem())
		e.Value = fieldVal.Elem().Interface()

		// Trigger NotifyPropertyChanged
		proto.PropertyChanged(tag.FieldName, &e)

		return nil
	})
}

// reflectArgs builds up reflect args
// We loop through the function arguments and use the types of each argument to decode the jsArgs
// If the function has more arguments than we have jsArgs, they're passed in as Zero values
// If the function has less arguments than jsArgs, the superfluous jsArgs are silently discarded
func reflectArgs(handler reflect.Method, proto Interface, jsArgs []*js.Object) []reflect.Value {
	reflectArgs := make([]reflect.Value, handler.Type.NumIn())
	reflectArgs[0] = reflect.ValueOf(proto)
	for i := 1; i < handler.Type.NumIn(); i++ {
		argType := handler.Type.In(i)
		var arg reflect.Value
		if argType.Kind() == reflect.Ptr {
			arg = reflect.New(argType.Elem())
			reflectArgs[i] = arg
		} else {
			arg = reflect.New(argType)
			reflectArgs[i] = arg.Elem()
		}

		if len(jsArgs) > i {
			if err := Decode(jsArgs[i-1], arg.Interface()); err != nil {
				panic(fmt.Sprintf("Error while decoding argument %v: %v", i, err))
			}
		}
	}

	return reflectArgs
}

func handlerCallback(handler reflect.Method) *js.Object {
	return js.MakeFunc(func(this *js.Object, jsArgs []*js.Object) interface{} {
		f := func() {
			handler.Func.Call(reflectArgs(handler, jsMap[this.Get(protoIndexKey).Int()], jsArgs))
		}

		// We delay this call until after other event processing, this avoids user callbacks being called before our own processing
		this.Call("async", f, 1)

		return nil
	})
}

func computeCallback(handler reflect.Method) *js.Object {
	return js.MakeFunc(func(this *js.Object, jsArgs []*js.Object) interface{} {
		returnArgs := handler.Func.Call(reflectArgs(handler, jsMap[this.Get(protoIndexKey).Int()], jsArgs))
		return returnArgs[0].Interface()
	})
}
