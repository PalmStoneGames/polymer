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
		// Create a new Go side object and keep it around in this closure
		// That way, we can keep track of it across callbacks and calls
		refVal := reflect.New(refType.Elem())
		proto := refVal.Interface().(Interface)
		refVal = refVal.Elem()

		// Set the proto value, this is needed because we get our callers to embed *polymer.Proto, and it needs to get instantiated
		refVal.FieldByName("Proto").Set(reflect.ValueOf(&Proto{}))

		// Store ourselves in js land so we can map js to proto
		jsMap = append(jsMap, proto)
		this.Set(protoIndexKey, len(jsMap)-1)

		// Set data on the proto
		proto.data().object = this
		proto.data().tags = tags

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
			if err := Decode(this.Get(getJsName(tag.FieldName)), refVal.Elem().Field(tag.FieldIndex).Addr().Interface()); err != nil {
				panic(fmt.Sprintf("Error while decoding polymer field value for %v: %v", tag.FieldName, err))
			}

			this.Call("addEventListener", getJsPropertyChangedEvent(tag.FieldName), propertyChangeCallback(refVal, tag))
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

func propertyChangeCallback(refVal reflect.Value, tag *fieldTag) func(*js.Object) {
	return func(jsEvent *js.Object) {
		// Decode the event
		var e PropertyChangedEvent
		if err := decodeStruct(jsEvent, &e); err != nil {
			panic(fmt.Sprintf("Error while decoding event: %v", err))
		}

		// Set the field on the Go side
		refVal.Elem().Field(tag.FieldIndex).Set(reflect.ValueOf(e.Value))
	}
}

func handlerCallback(handler reflect.Method) *js.Object {
	return js.MakeFunc(func(this *js.Object, jsArgs []*js.Object) interface{} {
		f := func() {
			// Lookup the proto
			proto := jsMap[this.Get(protoIndexKey).Int()]

			// Build up reflect args
			// We loop through the function arguments and use the types of each argument to decode the jsArgs
			// If the function has more arguments than we have jsArgs, they're passed in as Zero values
			// If the function has less arguments than jsArgs, the superfluous jsArgs are silently discarded
			reflectArgs := make([]reflect.Value, handler.Type.NumIn())
			reflectArgs[0] = reflect.ValueOf(proto)
			for i := 1; i < handler.Type.NumIn(); i++ {
				arg := reflect.New(handler.Type.In(i))
				reflectArgs[i] = arg.Elem()

				if len(jsArgs) > i {
					if err := Decode(jsArgs[i-1], arg.Interface()); err != nil {
						panic(fmt.Sprintf("Error while decoding argument %v: %v", i, err))
					}
				}
			}

			handler.Func.Call(reflectArgs)
		}

		// We delay this call until after other event processing, this avoids user callbacks being called before our own processing
		this.Call("async", f, 1)

		return nil
	})
}
