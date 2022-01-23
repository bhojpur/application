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

import "net/http"

// Global global role instance
var Global = &Role{}

// Register register role with conditions
func Register(name string, fc Checker) {
	Global.Register(name, fc)
}

// Allow allows permission mode for roles
func Allow(mode PermissionMode, roles ...string) *Permission {
	return Global.Allow(mode, roles...)
}

// Deny deny permission mode for roles
func Deny(mode PermissionMode, roles ...string) *Permission {
	return Global.Deny(mode, roles...)
}

// Get role defination
func Get(name string) (Checker, bool) {
	return Global.Get(name)
}

// Remove role definition from global role instance
func Remove(name string) {
	Global.Remove(name)
}

// Reset role definitions from global role instance
func Reset() {
	Global.Reset()
}

// MatchedRoles return defined roles from user
func MatchedRoles(req *http.Request, user interface{}) []string {
	return Global.MatchedRoles(req, user)
}

// HasRole check if current user has role
func HasRole(req *http.Request, user interface{}, roles ...string) bool {
	return Global.HasRole(req, user)
}

// NewPermission initialize a new permission for default role
func NewPermission() *Permission {
	return Global.NewPermission()
}
