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
	Type string    `polymer-decode:"type"`
	Time time.Time `polymer-decode:"timeStamp"`

	IsTrusted        bool `polymer-decode:"isTrusted"`
	Cancelable       bool `polymer-decode:"cancelable"`
	DefaultPrevented bool `polymer-decode:"defaultPrevented"`
	Bubbles          bool `polymer-decode:"bubbles"`
	CancelBubble     bool `polymer-decode:"cancelBubble"`

	SourceElement Element   `polymer-decode:"srcElement"`
	Target        Element   `polymer-decode:"target"`
	Path          []Element `polymer-decode:"path"`
}

type PropertyChangedEvent struct {
	Event
	JSValue *js.Object `polymer-decode:"detail.value"`
}
