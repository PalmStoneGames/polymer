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
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"reflect"
)

var domAPIConstructor *js.Object

// FlushDOM flushes pending changes to the DOM
// Insert, append, and remove operations are transacted lazily in certain cases for performance.
// In order to interrogate the DOM (for example, offsetHeight, getComputedStyle, etc.) immediately after one of these operations, call this function first.
func FlushDOM() {
	js.Global.Get("Polymer").Get("dom").Call("flush")
}

func WrapDOMElement(el dom.Element) Element {
	// Check if the element is already wrapped, if so, avoid double-wrapping it
	if isWrapped(el.Underlying()) {
		// Try and cast the dom.Element to a polymer.Element, if it works, just return that
		// Otherwise, wrap it in a PolymerWrappedElement directly without wrapping the js element, as it's already wrapped
		if newEl, ok := el.(Element); ok {
			return newEl
		} else {
			return &PolymerWrappedElement{el}
		}
	}

	return WrapJSElement(el.Underlying())
}

func WrapJSElement(obj *js.Object) Element {
	if obj == nil || obj == js.Undefined {
		return nil
	}

	if isWrapped(obj) {
		obj = obj.Get("node")
	}

	return &PolymerWrappedElement{dom.WrapElement(polymerDOM(obj))}
}

func polymerDOM(obj *js.Object) *js.Object {
	// Check if the element is already wrapped, if so, avoid double-wrapping it
	if isWrapped(obj) {
		return obj
	}

	return js.Global.Get("Polymer").Call("dom", obj)
}

func isWrapped(obj *js.Object) bool {
	if domAPIConstructor == nil || domAPIConstructor == js.Undefined {
		domAPIConstructor = js.Global.Get("Polymer").Get("DomApi").Get("constructor")
	}

	if domAPIConstructor == nil || domAPIConstructor == js.Undefined {
		panic("Polymer has not correctly initialized yet")
	}

	return obj.Get("constructor") == domAPIConstructor
}

func objToNodeSlice(obj *js.Object) []dom.Node {
	if obj == nil || obj == js.Undefined {
		return nil
	}

	nodes := make([]dom.Node, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		nodes[i] = dom.WrapNode(obj.Index(i))
	}

	return nodes
}

func objToElementSlice(obj *js.Object) []Element {
	if obj == nil || obj == js.Undefined {
		return nil
	}

	nodes := make([]Element, obj.Length())
	for i := 0; i < obj.Length(); i++ {
		nodes[i] = WrapJSElement(obj.Index(i))
	}

	return nodes
}

func GetWindow() dom.Window {
	return window{dom.GetWindow()}
}

type window struct {
	dom.Window
}

func (w window) Document() dom.Document {
	doc := w.Window.Document()
	return &document{
		Document:       doc,
		wrappedElement: WrapDOMElement(doc.DocumentElement()),
	}
}

type document struct {
	dom.Document
	wrappedElement dom.Element
}

func (d *document) CreateElement(name string) dom.Element {
	return WrapDOMElement(d.Document.CreateElement(name))
}

func (d *document) CreateElementNS(namespace, name string) dom.Element {
	return WrapDOMElement(d.Document.CreateElementNS(namespace, name))
}

func (d *document) ElementFromPoint(x, y int) dom.Element {
	return WrapDOMElement(d.Document.ElementFromPoint(x, y))
}

func (d *document) GetElementByID(id string) dom.Element {
	return d.wrappedElement.QuerySelector("#" + id)
}

func (d *document) GetElementsByClassName(name string) []dom.Element {
	return d.wrappedElement.QuerySelectorAll("." + name)
}

func (d *document) GetElementsByTagName(name string) []dom.Element {
	return d.wrappedElement.QuerySelectorAll(name)
}

func (d *document) GetElementsByTagNameNS(ns, name string) []dom.Element {
	panic("Operation not supported")
}

func (d *document) QuerySelector(sel string) dom.Element {
	return d.wrappedElement.QuerySelector(sel)
}

func (d *document) QuerySelectorAll(sel string) []dom.Element {
	return d.wrappedElement.QuerySelectorAll(sel)
}

type Element interface {
	dom.Element

	// Root returns the local DOM root of the current element
	Root() Element

	// GetDistributedNodes returns the nodes distributed to a <content> insertion point
	// only returns useful results when called on a <content> element
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
	GetDistributedNodes() []dom.Node
	// GetDestinationInsertionPoints returns the <content> nodes this element will be distributed to
	// only returns useful results when called on an element that’s being distributed.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
	GetDestinationInsertionPoints() []dom.Node
	// GetContentChildNodes accepts a css selector that points to a <content> node and returns all nodes that have been distributed to it
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
	GetContentChildNodes(selector string) []dom.Node
	// GetContentChildNodes accepts a css selector that points to a <content> node and returns all elements that have been distributed to it
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
	GetContentChildren(selector string) []Element

	// GetEffectiveChildNodes returns a list of effective child nodes for this element.
	// Effective child nodes are the child nodes of the element, with any insertion points replaced by their distributed child nodes.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
	GetEffectiveChildNodes() []dom.Node
	// GetEffectiveChildren returns a list of effective children for this element.
	// Effective children are the children of the element, with any insertion points replaced by their distributed children.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
	GetEffectiveChildren() []Element
	// QueryEffectiveChildren returns the first effective child that matches the selector
	// Effective children are the children of the element, with any insertion points replaced by their distributed children.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
	QueryEffectiveChildren(selector string) Element
	// QueryAllEffectiveChildren returns a slice of effective children that match selector
	// Effective children are the children of the element, with any insertion points replaced by their distributed children.
	// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
	QueryAllEffectiveChildren(selector string) []Element

	// ObserveNodes sets up an observer that will be called when new nodes are added or removed from this Element
	// ObserveNodes  behaves slightly differently depending on the node being observed:
	// - If the node being observed is a content node, the callback is called when the content node’s distributed children change.
	// - For any other node, the callback is called when the node’s effective children change.
	ObserveNodes(func(*ObservationInfo)) *Observer
	// UnobserveNodes stops an observer from receiving notifications
	UnobserveNodes(*Observer)

	// SubscribeEvent subscribes to an event using the passed callback
	// The callback may be strongly typed, the types will be automatically decoded
	SubscribeEvent(event string, callback interface{}) *EventSubscribtion
}

type PolymerWrappedElement struct {
	dom.Element
}

func (el *PolymerWrappedElement) GetElementsByClassName(name string) []dom.Element {
	return el.QuerySelectorAll("." + name)
}

func (el *PolymerWrappedElement) GetElementsByTagName(name string) []dom.Element {
	return el.QuerySelectorAll(name)
}

func (el *PolymerWrappedElement) GetElementsByTagNameNS(ns, name string) []dom.Element {
	panic("Operation not supported")
}

func (el *PolymerWrappedElement) AppendChild(node dom.Node) {
	obj := node.Underlying()
	if obj.Get("constructor") == js.Global.Get("Polymer").Get("DomApi").Get("constructor") {
		obj = obj.Get("node")
	}

	el.Element.AppendChild(dom.WrapElement(obj))
}

type EventSubscribtion struct {
	event   string
	funcObj *js.Object
}

func (el *PolymerWrappedElement) SubscribeEvent(event string, callback interface{}) *EventSubscribtion {
	refVal := reflect.ValueOf(callback)
	var funcObj *js.Object
	switch refVal.Kind() {
	case reflect.Func:
		funcObj = eventHandlerCallback(refVal)
	case reflect.Chan:
		funcObj = eventChanCallback(refVal)
	default:
		panic(fmt.Sprint("Expected callback of kind %s or %s, but got %s", reflect.Func, reflect.Chan, refVal.Kind()))
	}

	sub := &EventSubscribtion{
		event:   event,
		funcObj: funcObj,
	}

	el.Underlying().Get("node").Call("addEventListener", event, sub.funcObj)
	return sub
}

func (el *PolymerWrappedElement) UnsubscribeEvent(sub *EventSubscribtion) {
	el.Underlying().Call("removeEventListener", sub.event, sub.funcObj)
}

// Root returns the local DOM root of the current element
func (el *PolymerWrappedElement) Root() Element {
	// root is set on the polymer element, but not on its wrapped equivalent, so drill through the wrapper to get the root
	return WrapJSElement(el.Underlying().Get("node").Get("root"))
}

// ObservationInfo is the structure used to hand data to ObserveNodes callbacks
type ObservationInfo struct {
	Observer *Observer

	AddedNodes, RemovedNodes []dom.Node
}

// Observer is the structure used to track an observation using ObserveNodes/UnobserveNodes
type Observer struct {
	Element Element
	object  *js.Object
}

// GetDistributedNodes returns the nodes distributed to a <content> insertion point
// only returns useful results when called on a <content> element
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
func (el *PolymerWrappedElement) GetDistributedNodes() []dom.Node {
	return objToNodeSlice(el.Underlying().Call("getDistributedNodes"))
}

// GetDestinationInsertionPoints returns the <content> nodes this element will be distributed to
// only returns useful results when called on an element that’s being distributed.
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
func (el *PolymerWrappedElement) GetDestinationInsertionPoints() []dom.Node {
	return objToNodeSlice(el.Underlying().Call("getDestinationInsertionPoints"))
}

// GetContentChildNodes accepts a css selector that points to a <content> node and returns all nodes that have been distributed to it
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
func (el *PolymerWrappedElement) GetContentChildNodes(selector string) []dom.Node {
	return objToNodeSlice(el.Underlying().Call("getContentChildNodes"))
}

// GetContentChildNodes accepts a css selector that points to a <content> node and returns all elements that have been distributed to it
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#distributed-children
func (el *PolymerWrappedElement) GetContentChildren(selector string) []Element {
	return objToElementSlice(el.Underlying().Call("getContentChildren"))
}

// GetEffectiveChildNodes returns a list of effective child nodes for this element.
// Effective child nodes are the child nodes of the element, with any insertion points replaced by their distributed child nodes.
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
func (el *PolymerWrappedElement) GetEffectiveChildNodes() []dom.Node {
	return objToNodeSlice(el.Underlying().Call("getEffectiveChildNodes"))
}

// GetEffectiveChildren returns a list of effective children for this element.
// Effective children are the children of the element, with any insertion points replaced by their distributed children.
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
func (el *PolymerWrappedElement) GetEffectiveChildren() []Element {
	return objToElementSlice(el.Underlying().Call("getEffectiveChildren"))
}

// QueryEffectiveChildren returns the first effective child that matches the selector
// Effective children are the children of the element, with any insertion points replaced by their distributed children.
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
func (el *PolymerWrappedElement) QueryEffectiveChildren(selector string) Element {
	return WrapJSElement(el.Underlying().Call("queryEffectiveChildren"))
}

// QueryAllEffectiveChildren returns a slice of effective children that match selector
// Effective children are the children of the element, with any insertion points replaced by their distributed children.
// Details can be found at https://www.polymer-project.org/1.0/docs/devguide/local-dom.html#effective-children
func (el *PolymerWrappedElement) QueryAllEffectiveChildren(selector string) []Element {
	return objToElementSlice(el.Underlying().Call("queryAllEffectiveChildren"))
}

// ObserveNodes sets up an observer that will be called when new nodes are added or removed from this Element
// ObserveNodes  behaves slightly differently depending on the node being observed:
// - If the node being observed is a content node, the callback is called when the content node’s distributed children change.
// - For any other node, the callback is called when the node’s effective children change.
func (el *PolymerWrappedElement) ObserveNodes(f func(*ObservationInfo)) *Observer {
	obs := &Observer{}
	obs.Element = el
	obs.object = el.Underlying().Call("observeNodes", wrapObserveNodesCallback(obs, f))
	return obs
}

func wrapObserveNodesCallback(obs *Observer, f func(*ObservationInfo)) func(*js.Object) {
	return func(obj *js.Object) {
		info := ObservationInfo{
			Observer:     obs,
			AddedNodes:   objToNodeSlice(obj.Get("addedNodes")),
			RemovedNodes: objToNodeSlice(obj.Get("removedNodes")),
		}

		f(&info)
	}
}

// UnobserveNodes stops an observer from receiving notifications
func (el *PolymerWrappedElement) UnobserveNodes(obs *Observer) {
	el.Underlying().Call("unobserveNodes", obs.object)
}
