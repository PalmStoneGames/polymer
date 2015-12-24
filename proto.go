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
	"github.com/gopherjs/gopherjs/js"
)

// Interface represents the interface implemented by all type prototypes
// Any type implementing this interface can be registered with polymer.Register()
// Most of this interface can be implemented by embedded polymer.Proto
// The notable exception to this is TagName, which must always be manually implemented
type Interface interface {
	// TagName should return the name of the tag this type will be handling
	TagName() string
	// Extends should return the name of the tag that this element extends.
	// Currently, polymer only supports extending default elements, not user-defined ones
	Extends() string

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

	// PropertyChanged is called when polymer detects a property change on a property
	PropertyChanged(fieldName string, e *PropertyChangedEvent)

	// Internal utility
	data() *Proto
}

// Proto represents a prototype for a polymer type
// it's meant to be embedded by the structures used to implements polymer tags
type Proto struct {
	this *js.Object
	Element
	tags []*fieldTag
}

func (p *Proto) Extends() string { return "" }

func (p *Proto) Created()  {}
func (p *Proto) Ready()    {}
func (p *Proto) Attached() {}
func (p *Proto) Detached() {}

func (p *Proto) PropertyChanged(fieldName string, e *PropertyChangedEvent) {}

func (p *Proto) data() *Proto { return p }

// Notify notifies polymer that a value has changed
// TODO: Change Notify to accept a pointer to the field that changed instead of a path and a value, we're waiting on https://github.com/gopherjs/gopherjs/issues/364 for this
func (p *Proto) Notify(path string, val interface{}) {
	p.this.Call("set", path, val)
}
