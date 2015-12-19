package main

import (
	"fmt"

	"code.palmstonegames.com/polymer"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	polymer.Register(&NameTag{})
}

type NameTag struct {
	*polymer.Proto

	Name string `polymer:"bind"`
}

func (n *NameTag) TagName() string {
	return "name-tag"
}

func (n *NameTag) Created() {
	n.Name = "Alice"
}

func (n *NameTag) HandleNameChange(map[string]interface{}, *js.Object) {
	fmt.Printf("Go: Name = %v\n", n.Name)
}

func main() {}
