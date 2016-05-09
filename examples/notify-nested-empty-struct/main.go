package main

import (
	"fmt"

	"code.palmstonegames.com/polymer"
)

func init() {
	polymer.Register("data-container", &DataContainer{})
}

type Account struct {
	Login string `polymer:"bind"`
	Email string `polymer:"bind"`
}

type DataContainer struct {
	*polymer.Proto
	Account Account `polymer:"bind"`
}

func (d *DataContainer) Created() {
	//The Account structure will be initialized empty.
}

// Ready is a callback
func (d *DataContainer) Ready() {
	// We need to call the notify on the empty structure otherwise the JS->Go binding is not done.
	d.Notify("account")
}

func (d *DataContainer) HandleInput() {
	polymer.Async(1, func() {
		fmt.Printf("Account: %#v\n", d.Account)
	})
}

func main() {}
