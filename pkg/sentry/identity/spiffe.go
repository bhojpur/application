package identity

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
	"strings"

	"github.com/pkg/errors"
)

// CreateSPIFFEID returns a SPIFFE standard unique id for the given trust domain, namespace and appID.
func CreateSPIFFEID(trustDomain, namespace, appID string) (string, error) {
	if trustDomain == "" {
		return "", errors.New("can't create spiffe id: trust domain is empty")
	}
	if namespace == "" {
		return "", errors.New("can't create spiffe id: namespace is empty")
	}
	if appID == "" {
		return "", errors.New("can't create spiffe id: app id is empty")
	}

	// Validate according to the SPIFFE spec
	if strings.Contains(trustDomain, ":") {
		return "", errors.New("trust domain cannot contain the : character")
	}
	if len([]byte(trustDomain)) > 255 {
		return "", errors.New("trust domain cannot exceed 255 bytes")
	}

	id := fmt.Sprintf("spiffe://%s/ns/%s/%s", trustDomain, namespace, appID)
	if len([]byte(id)) > 2048 {
		return "", errors.New("spiffe id cannot exceed 2048 bytes")
	}
	return id, nil
}
