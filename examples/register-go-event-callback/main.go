package main

import (
	"code.palmstonegames.com/polymer"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	polymer.Register("fancy-square", &FancySquare{})
}

type FancySquare struct {
	*polymer.Proto

	// Vars
	ID   int64  `polymer:"bind"`
	Name string `polymer:"bind"`
}

func (s *FancySquare) Ready() {
	s.SubscribeEvent("click", func(e *polymer.Event) { js.Global.Get("console").Call("log", e) })
}

func main() {}
