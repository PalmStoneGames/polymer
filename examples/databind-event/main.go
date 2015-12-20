package main

import (
	"fmt"
	"math/rand"

	"code.palmstonegames.com/polymer"
)

func init() {
	polymer.Register(&NameTag{})
}

type NameTag struct {
	*polymer.Proto

	ID   int64  `polymer:"bind"`
	Name string `polymer:"bind"`
}

func (n *NameTag) TagName() string {
	return "name-tag"
}

func (n *NameTag) Ready() {
	n.ID = rand.Int63()
	fmt.Printf("%v: Initial Name = %v\n", n.ID, n.Name)
}

func (n *NameTag) HandleNameChange(e polymer.Event) {
	fmt.Printf("%v: Name = %v\n", n.ID, n.Name)
}

func main() {}
