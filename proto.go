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
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// Interface represents the interface implemented by all type prototypes
// Any type implementing this interface can be registered with polymer.Register()
// Most of this interface can be implemented by embedded polymer.Proto
// The notable exception to this is TagName, which must always be manually implemented
type Interface interface {
	// Basic info
	TagName() string
	Extends() string

	// Lifetime callbacks
	Created()
	Ready()
	Attached()
	Detached()

	// Internal utility
	data() *Proto
}

type fieldTag struct {
	FieldIndex int
	FieldName  string
	Bind       bool
}

// Proto represents a prototype for a polymer type
// it's meant to be embedded by the structures used to implements polymer tags
type Proto struct {
	object *js.Object
	tags   []*fieldTag
}

func (p *Proto) Extends() string { return "" }
func (p *Proto) Created()        {}
func (p *Proto) Ready()          {}
func (p *Proto) Attached()       {}
func (p *Proto) Detached()       {}
func (p *Proto) data() *Proto    { return p }

func parseTags(refType reflect.Type) []*fieldTag {
	var tags []*fieldTag
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		tagText := field.Tag.Get("polymer")
		if tagText == "" {
			continue
		}
		tag := strings.Split(tagText, ",")

		f := fieldTag{
			FieldIndex: i,
			FieldName:  field.Name,
		}

		for i := 1; i < len(tag); i++ {
			switch tag[i] {
			case "bind":
				f.Bind = true
			}
		}

		tags = append(tags, &f)
	}

	return tags
}

func parseHandlers(refType reflect.Type) []reflect.Method {
	var handlers []reflect.Method

	for i := 0; i < refType.NumMethod(); i++ {
		method := refType.Method(i)

		if strings.HasPrefix(method.Name, "Handle") {
			handlers = append(handlers, method)
		}
	}

	return handlers
}
