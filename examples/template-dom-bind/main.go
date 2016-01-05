package main

import (
	"time"

	"code.palmstonegames.com/polymer"
)

var data = &Data{Text: "This text is set from Go", RemainingTime: 5}

type Data struct {
	*polymer.BindProto

	Text          string
	RemainingTime int
	Clock         time.Time
	Click         chan *polymer.Event `polymer:"handler"`
}

func (d *Data) ComputeTime() string {
	return d.Clock.String()
}

func (d *Data) HandleClick(e *polymer.Event) {
	polymer.Log("Handleclick triggered: ", e)
}

func main() {
	// Wait until polymer has finished initializing
	<-polymer.OnReady()

	// Find #tmpl and bind it
	polymer.GetDocument().GetElementByID("tmpl").(*polymer.AutoBindGoTemplate).Bind(data)

	// Start a goroutine to dynamically update the timer every second and notify
	go func() {
		for data.RemainingTime > 0 {
			time.Sleep(time.Second)
			data.RemainingTime--
			data.Notify("RemainingTime", data.RemainingTime)
		}

		data.Text = "This text is ALSO set from Go, in a goroutine, after 5 seconds"
		data.Notify("Text", data.Text)
	}()

	// Start a 2nd goroutine to update the live clock, this code is identical to the computed-property example
	go func() {
		for {
			// Set the clock
			data.Clock = time.Now()

			// Notify
			data.Notify("Clock", data.Clock)

			// Wait
			time.Sleep(time.Millisecond * 100)
		}
	}()

	// Start a 3rd goroutine to listen on the Click channel
	go func() {
		for e := range data.Click {
			polymer.Log("click triggered: ", e)
		}
	}()
}
