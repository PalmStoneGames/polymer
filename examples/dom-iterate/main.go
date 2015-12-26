package main

import (
	"code.palmstonegames.com/polymer"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	polymer.Register(&FancyInput{})
	polymer.Register(&InputList{})
}

type FancyInput struct {
	*polymer.Proto
	Value string `polymer:"bind"`
}

func (i *FancyInput) TagName() string {
	return "fancy-input"
}

type InputList struct {
	*polymer.Proto
}

func (t *InputList) TagName() string {
	return "input-list"
}

func (t *InputList) Ready() {
	// Print out inputs that we can find on the root
	// This shouldn't find any due to the shady DOM api (correctly) considering them part of the local DOM of fancy-input
	js.Global.Get("console").Call("log", "Printing inputs through t.Root().QuerySelectorAll()")
	js.Global.Get("console").Call("log", t.Root().QuerySelectorAll("input"))

	// Try and reach through the document instead
	// This shouldn't find any due to the shady DOM api (correctly) considering them part of the local DOM of fancy-input
	js.Global.Get("console").Call("log", "Printing inputs through polymer.GetWindow().Document()")
	js.Global.Get("console").Call("log", polymer.WrapDOMElement(polymer.GetWindow().Document().QuerySelector("input-list")).Root().QuerySelectorAll("input"))

	// Next, find fancy-inputs instead of the wrapped inputs, those we should be able to find
	js.Global.Get("console").Call("log", "Printing fancy-inputs through t.Root().QuerySelectorAll()")
	js.Global.Get("console").Call("log", t.Root().QuerySelectorAll("fancy-input"))

	// Next, drill all the way down to the inputs by going into fancy-input's local DOM
	js.Global.Get("console").Call("log", "Printing input element in nested shadow DOM")
	js.Global.Get("console").Call("log", polymer.WrapDOMElement(polymer.WrapDOMElement(polymer.GetWindow().Document().QuerySelector("input-list")).Root().QuerySelector("fancy-input")).Root().QuerySelector("input"))
}

func main() {}
