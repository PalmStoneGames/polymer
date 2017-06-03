package main

import (
	"time"

	"github.com/PalmStoneGames/polymer"
)

func init() {
	polymer.Register("tick-timer", &Timer{})
}

type Timer struct {
	*polymer.Proto

	H int `polymer:"bind"`
	M int `polymer:"bind"`
	S int `polymer:"bind"`
}

func (t *Timer) Created() {
	go func() {
		for {
			// Set the clock
			now := time.Now()
			t.H = now.Hour()
			t.M = now.Minute()
			t.S = now.Second()

			// Notify
			t.Notify("h")
			t.Notify("m")
			t.Notify("s")

			// Wait
			time.Sleep(time.Second)
		}
	}()
}

func main() {}
