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
}

// Proto represents a prototype for a polymer type
// It is initially empty, it's meant to be embedded by the structures used to implements polymer tags
type Proto struct{}

func (p Proto) Extends() string { return "" }
func (p Proto) Created()        {}
func (p Proto) Ready()          {}
func (p Proto) Attached()       {}
func (p Proto) Detached()       {}
