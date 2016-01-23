/*
Copyright 2015 Palm Stone Games, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package polymer

import (
	"time"

	"github.com/gopherjs/gopherjs/js"
)

type Event struct {
	Underlying *js.Object `polymer-decode:"underlying"`

	Type string    `polymer-decode:"event.type"`
	Time time.Time `polymer-decode:"event.timeStamp"`

	IsTrusted        bool `polymer-decode:"event.isTrusted"`
	Cancelable       bool `polymer-decode:"event.cancelable"`
	DefaultPrevented bool `polymer-decode:"event.defaultPrevented"`
	Bubbles          bool `polymer-decode:"event.bubbles"`
	CancelBubble     bool `polymer-decode:"event.cancelBubble"`

	LocalTarget Element   `polymer-decode:"localTarget"`
	RootTarget  Element   `polymer-decode:"rootTarget"`
	Path        []Element `polymer-decode:"path"`
}

type PropertyChangedEvent struct {
	Event
	JSValue *js.Object `polymer-decode:"event.detail.value"`
}

type MouseEvent struct {
	Event

	MovementX int `polymer-decode:"event.movementX"`
	MovementY int `polymer-decode:"event.movementY"`
	OffsetX   int `polymer-decode:"event.offsetX"`
	OffsetY   int `polymer-decode:"event.offsetY"`
	PageX     int `polymer-decode:"event.pageX"`
	PageY     int `polymer-decode:"event.pageY"`

	FromElement Element `polymer-decode:"event.fromElement"`
	ToElement   Element `polymer-decode:"event.toElement"`
}

func (e *Event) StopPropagation() {
	e.CancelBubble = true
	e.Underlying.Get("event").Call("stopPropagation")
}

func (e *Event) PreventDefault() {
	e.DefaultPrevented = true
	e.Underlying.Get("event").Call("preventDefault")
}
