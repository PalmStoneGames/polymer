package main

import (
	"strconv"

	"code.palmstonegames.com/polymer"
	"honnef.co/go/js/dom"
)

func init() {
	polymer.Register(&TableLayout{})
}

type TableLayout struct {
	*polymer.Proto
}

func (t *TableLayout) TagName() string {
	return "table-layout"
}

func (t *TableLayout) Ready() {
	document := dom.GetWindow().Document()
	shadowRoot := t.Root()
	for i := 1; i <= 10; i++ {
		el := document.CreateElement("div")
		el.SetTextContent(strconv.FormatInt(int64(i), 10))
		shadowRoot.AppendChild(el)
	}
}

func main() {}
