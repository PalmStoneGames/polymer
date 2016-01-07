package main

import (
	"code.palmstonegames.com/polymer"
	"time"
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
			t.Notify("H", t.H)
			t.Notify("M", t.M)
			t.Notify("S", t.S)

			// Wait
			time.Sleep(time.Second)
		}
	}()
}

func main() {}
