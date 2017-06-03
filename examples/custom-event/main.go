package main

import (
	"fmt"

	"github.com/PalmStoneGames/polymer"
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
			case e := <-p.DoCustomEvent:
				p.WasFired = "Yes"
				p.Notify("wasFired")
				fmt.Printf("%v\n", e.Underlying.Get("event").Get("detail").Get("message").String())
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
		if d.FirePassword == "Fire" {
			d.Fire("custom-event", map[string]interface{}{
				"message": "Event Fired from Child",
			})
		}
	})
}

func main() {}
