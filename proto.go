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
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// Interface represents the interface implemented by all type prototypes
// Any type implementing this interface can be registered with polymer.Register()
// Most of this interface can be implemented by embedded polymer.Proto
// The notable exception to this is TagName, which must always be manually implemented
type Interface interface {
	// Created is part of the lifecycle callbacks
	// It is called when the element is initially created, before the constructor
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/registering-elements.html#lifecycle-callbacks
	Created()

	// Ready is part of the lifecycle callbacks
	// The ready callback is called when an element’s local DOM is ready.
	// It is called after the element’s template has been stamped and all elements inside the element’s local DOM have been configured
	// (with values bound from parents, deserialized attributes, or else default values) and had their ready method called.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/registering-elements.html#ready-method
	Ready()

	// Attached is part of the lifecycle callbacks
	// It is called when the element is attached to the DOM
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/registering-elements.html#lifecycle-callbacks
	Attached()

	// Detached is part of the lifecycle callbacks
	// It is called when the element is detached from the DOM
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/registering-elements.html#lifecycle-callbacks
	Detached()

	// Internal utility
	data() *Proto
}

// Proto represents a prototype for a polymer type
// it's meant to be embedded by the structures used to implements polymer tags
type Proto struct {
	this *js.Object
	Element
	ready bool
}

func (p *Proto) Extends() string { return "" }

func (p *Proto) Created()  {}
func (p *Proto) Ready()    {}
func (p *Proto) Attached() {}
func (p *Proto) Detached() {}

func (p *Proto) data() *Proto { return p }

// This returns the underlying js object corresponding to the `this` magic value in javascript
// Unlike Underlying(), this object is not wrapped by Polymer.dom()
func (p *Proto) This() *js.Object { return p.this }

// Notify notifies polymer that a value has changed
func (p *Proto) Notify(path string) {
	refVal := getRefValForPath(lookupProto(p.this), strings.Split(path, "."))
	jsObj, _ := encodeRaw(refVal)
	p.doNotify(path, jsObj)
}

func (p *Proto) doNotify(path string, val interface{}) {
	p.this.Call("set", path, val)
}

func (p *Proto) Fire(event string, val interface{}) {
	p.this.Call("fire", event, val)
}

type AsyncHandle struct {
	jsHandle *js.Object
}

// Async calls the given callback asynchronously.
// If the specified wait time is -1, the callback will be ran with microtask timing (after the current method finishes, but before the next event from the event queue is processed)
// Otherwise, its ran waitTime milliseconds in the future. A waitTime of 1 can be useful to run a callback after all events currently in the queue have been processed.
// Returns a handle that can be used to cancel the task
func (p *Proto) Async(waitTime int, f func()) *AsyncHandle {
	handle := &AsyncHandle{}
	if waitTime == -1 {
		handle.jsHandle = p.this.Call("async", f)
	} else {
		handle.jsHandle = p.this.Call("async", f, waitTime)
	}

	return handle
}

func (p *Proto) CancelAsync(handle *AsyncHandle) {
	p.this.Call("cancelAsync", handle.jsHandle)
}
