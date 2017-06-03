package main

import (
	"strconv"

	"github.com/PalmStoneGames/polymer"
)

func init() {
	polymer.Register("table-layout", &TableLayout{})
}

type TableLayout struct {
	*polymer.Proto
}

func (t *TableLayout) Ready() {
	document := polymer.GetWindow().Document()
	shadowRoot := t.Root()
	for i := 1; i <= 10; i++ {
		el := document.CreateElement("div")
		el.SetTextContent(strconv.FormatInt(int64(i), 10))
		shadowRoot.AppendChild(el)
	}
}

func main() {}
