package main

import (
	"fmt"
	"time"

	"github.com/PalmStoneGames/polymer"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	polymer.Register("time-edit", &TimeEdit{})
}

// Note: This example is not idiomatic
// the polymer lib provides a polymer.Time type that already decodes and encodes to the correct format
// This example is purely for showing the usage of encode and decode
// for plugging time.Time into datetime or datetime-local fields, polymer.Time should be used
type CustomTime time.Time

func (c CustomTime) Encode() (*js.Object, bool) {
	return polymer.InterfaceToJsObject(time.Time(c).Format("2006-01-02T15:04:05")), !time.Time(c).IsZero()
}

func (c *CustomTime) Decode(val *js.Object) error {
	t, err := time.Parse("2006-01-02T15:04:05", val.String())
	if err != nil {
		return err
	}

	*c = CustomTime(t)
	return nil
}

type TimeEdit struct {
	*polymer.Proto

	Time CustomTime `polymer:"bind"`
}

func (edit *TimeEdit) Created() {
	edit.Time = CustomTime(time.Now())
}

func (edit *TimeEdit) Ready() {
	c := make(chan *polymer.Event)
	edit.SubscribeEvent("time-changed", c)

	go func() {
		for {
			<-c
			fmt.Println(time.Time(edit.Time).String())
		}
	}()
}

func main() {}
