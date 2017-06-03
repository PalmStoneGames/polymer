package main

import (
	"github.com/PalmStoneGames/polymer"
)

func init() {
	polymer.Register("el-one", &ElOne{})
	polymer.Register("el-two", &ElTwo{})
}

type ElOne struct {
	*polymer.Proto

	Data string `polymer:"bind"`
}

type ElTwo struct {
	*polymer.Proto

	Text string `polymer:"bind"`
}

func (el *ElOne) Created() {
	el.Data = "This is a test string set from Go"
}

func main() {}
