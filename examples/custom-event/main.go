package main

import (
	"fmt"

	"code.palmstonegames.com/polymer"
)

func init() {
	polymer.Register("parent-container", &ParentContainer{})
	polymer.Register("data-container", &DataContainer{})
}

type ParentContainer struct {
	*polymer.Proto

	DoCustomEvent chan *polymer.Event `polymer:"handler"`

	WasFired string `polymer:"bind"`
}

func (p *ParentContainer) Created() {
	p.WasFired = "Not Yet"
}

// Ready is a callback
func (p *ParentContainer) Ready() {
	p.listenToggleEvents()
}

func (p *ParentContainer) listenToggleEvents() {
	go func() {
		for {
			select {
			case <-p.DoCustomEvent:
				p.WasFired = "Yes"
				p.Notify("wasFired")
			}

		}
	}()
}

type DataContainer struct {
	*polymer.Proto

	FirePassword string `polymer:"bind"`
}

func (d *DataContainer) Created() {
}

// Ready is a callback
func (d *DataContainer) Ready() {
	d.Notify("firePassword")
}

func (d *DataContainer) HandleInput() {
	polymer.Async(1, func() {
		fmt.Printf("%s\n", d.FirePassword)
		if d.FirePassword == "Fire" {
			d.Fire("custom-event", nil)
		}
	})
}

func main() {}
