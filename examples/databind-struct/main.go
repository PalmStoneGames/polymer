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

	TestStruct TestStruct `polymer:"bind"`
}

type TestStruct struct {
	Time time.Time
}

func (t *Timer) Created() {
	go func() {
		for {
			// Set the clock
			t.TestStruct.Time = time.Now()

			// Notify
			t.Notify("testStruct")

			// Wait
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (t *Timer) ComputeTime(testStruct *TestStruct) string {
	return testStruct.Time.String()
}

func main() {}
