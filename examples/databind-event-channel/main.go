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

	ID         int64               `polymer:"bind"`
	Name       string              `polymer:"bind"`
	NameChange chan *polymer.Event `polymer:"handler"`
}

func (n *NameTag) TagName() string {
	return "name-tag"
}

func (n *NameTag) Created() {
	n.ID = rand.Int63()

	// Startup event listeners
	go func() {
		for _ = range n.NameChange {
			fmt.Printf("%v: HandleNameChange event. Name = %v\n", n.ID, n.Name)
		}
	}()
}

func (n *NameTag) Ready() {
	fmt.Printf("%v: Initial Name = %v\n", n.ID, n.Name)
}

func main() {}
