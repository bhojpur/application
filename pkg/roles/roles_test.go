package roles_test

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
	"testing"

	"github.com/bhojpur/application/pkg/roles"
)

func TestAllow(t *testing.T) {
	permission := roles.Allow(roles.Read, "api")

	if !permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has no permission to Update")
	}

	if permission.HasPermission(roles.Read, "admin") {
		t.Errorf("admin should has no permission to Read")
	}

	if permission.HasPermission(roles.Update, "admin") {
		t.Errorf("admin should has no permission to Update")
	}
}

func TestDeny(t *testing.T) {
	permission := roles.Deny(roles.Create, "api")

	if !permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has permission to Update")
	}

	if permission.HasPermission(roles.Create, "api") {
		t.Errorf("API should has no permission to Update")
	}

	if !permission.HasPermission(roles.Read, "admin") {
		t.Errorf("admin should has permission to Read")
	}

	if !permission.HasPermission(roles.Create, "admin") {
		t.Errorf("admin should has permission to Update")
	}
}

func TestCRUD(t *testing.T) {
	permission := roles.Allow(roles.CRUD, "admin")
	if !permission.HasPermission(roles.Read, "admin") {
		t.Errorf("Admin should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, "admin") {
		t.Errorf("Admin should has permission to Update")
	}

	if permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has no permission to Read")
	}

	if permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has no permission to Update")
	}
}

func TestAll(t *testing.T) {
	permission := roles.Allow(roles.Update, roles.Anyone)

	if permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has no permission to Read")
	}

	if !permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has permission to Update")
	}

	permission2 := roles.Deny(roles.Update, roles.Anyone)

	if !permission2.HasPermission(roles.Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if permission2.HasPermission(roles.Update, "api") {
		t.Errorf("API should has no permission to Update")
	}
}

func TestCustomizePermission(t *testing.T) {
	var customized roles.PermissionMode = "customized"
	permission := roles.Allow(customized, "admin")

	if !permission.HasPermission(customized, "admin") {
		t.Errorf("Admin should has customized permission")
	}

	if permission.HasPermission(roles.Read, "admin") {
		t.Errorf("Admin should has no permission to Read")
	}

	permission2 := roles.Deny(customized, "admin")

	if permission2.HasPermission(customized, "admin") {
		t.Errorf("Admin should has customized permission")
	}

	if !permission2.HasPermission(roles.Read, "admin") {
		t.Errorf("Admin should has no permission to Read")
	}
}
