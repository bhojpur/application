package roles

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"net/http"
)

const (
	// Anyone is a role for any one
	Anyone = "*"
)

// Checker check current request match this role or not
type Checker func(req *http.Request, user interface{}) bool

// New initialize a new `Role`
func New() *Role {
	return &Role{}
}

// Role is a struct contains all roles definitions
type Role struct {
	definitions map[string]Checker
}

// Register register role with conditions
func (role *Role) Register(name string, fc Checker) {
	if role.definitions == nil {
		role.definitions = map[string]Checker{}
	}

	definition := role.definitions[name]
	if definition != nil {
		fmt.Printf("Role `%v` already defined, overwrited it!\n", name)
	}
	role.definitions[name] = fc
}

// NewPermission initialize permission
func (role *Role) NewPermission() *Permission {
	return &Permission{
		Role:         role,
		AllowedRoles: map[PermissionMode][]string{},
		DeniedRoles:  map[PermissionMode][]string{},
	}
}

// Allow allows permission mode for roles
func (role *Role) Allow(mode PermissionMode, roles ...string) *Permission {
	return role.NewPermission().Allow(mode, roles...)
}

// Deny deny permission mode for roles
func (role *Role) Deny(mode PermissionMode, roles ...string) *Permission {
	return role.NewPermission().Deny(mode, roles...)
}

// Get role defination
func (role *Role) Get(name string) (Checker, bool) {
	fc, ok := role.definitions[name]
	return fc, ok
}

// Remove role definition
func (role *Role) Remove(name string) {
	delete(role.definitions, name)
}

// Reset role definitions
func (role *Role) Reset() {
	role.definitions = map[string]Checker{}
}

// MatchedRoles return defined roles from user
func (role *Role) MatchedRoles(req *http.Request, user interface{}) (roles []string) {
	if definitions := role.definitions; definitions != nil {
		for name, definition := range definitions {
			if definition(req, user) {
				roles = append(roles, name)
			}
		}
	}
	return
}

// HasRole check if current user has role
func (role *Role) HasRole(req *http.Request, user interface{}, roles ...string) bool {
	if definitions := role.definitions; definitions != nil {
		for _, name := range roles {
			if definition, ok := definitions[name]; ok {
				if definition(req, user) {
					return true
				}
			}
		}
	}
	return false
}
