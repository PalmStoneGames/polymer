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
			this.Set(getJsName(tag.FieldName), refVal.Elem().Field(tag.FieldIndex).Interface())
			this.Call("addEventListener", getPropertyChangedEventName(tag.FieldName), propertyChangeCallback(refVal, tag))
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
	return func(e *js.Object) {
		refVal.Elem().Field(tag.FieldIndex).Set(reflect.ValueOf(e.Get("detail").Get("value").Interface()))
	}
}

func handlerCallback(handler reflect.Method) *js.Object {
	return js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
		f := func() {
			// Lookup the proto
			proto := jsMap[this.Get(protoIndexKey).Int()]

			// Build up reflect args
			reflectArgs := make([]reflect.Value, len(args)+1)
			reflectArgs[0] = reflect.ValueOf(proto)
			for i, arg := range args {
				reflectArgs[i+1] = reflect.ValueOf(arg.Interface())
			}

			// Do the call, add a defer to recover from failures and give a more useful error message
			defer func() {
				if recover() != nil {
					errStr := fmt.Sprintf("Expected %v to have %v arguments", handler.Name, len(reflectArgs)-1)
					for i := 0; i < len(args); i++ {
						errStr += fmt.Sprintf("\n%v: %T", i+1, args[i].Interface())
					}
					panic(errStr)
				}
			}()

			handler.Func.Call(reflectArgs)
		}

		// We delay this call until after other event processing, this avoids user callbacks being called before our own processing
		this.Call("async", f, 1)

		return nil
	})
}
