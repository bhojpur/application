package validations

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
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// The consts and vars beginning with dns* were taken from:
// https://github.com/kubernetes/apimachinery/blob/fc49b38c19f02a58ebc476347e622142f19820b9/pkg/util/validation/validation.go
const (
	dns1123LabelFmt       string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	dns1123LabelErrMsg    string = "a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"
	dns1123LabelMaxLength int    = 63
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// ValidateKubernetesAppID returns a bool that indicates whether a Bhojpur Application
// app id is valid for the Kubernetes platform.
func ValidateKubernetesAppID(appID string) error {
	if appID == "" {
		return errors.New("value for the bhojpur.net/app-id annotation is empty")
	}
	r := isDNS1123Label(serviceName(appID))
	if len(r) == 0 {
		return nil
	}
	s := fmt.Sprintf("invalid app id(input: %s, service: %s): %s", appID, serviceName(appID), strings.Join(r, ","))
	return errors.New(s)
}

func serviceName(appID string) string {
	return fmt.Sprintf("%s-app", appID)
}

// The function was taken as-is from:
// https://github.com/kubernetes/apimachinery/blob/fc49b38c19f02a58ebc476347e622142f19820b9/pkg/util/validation/validation.go
func isDNS1123Label(value string) []string {
	var errs []string
	if len(value) > dns1123LabelMaxLength {
		errs = append(errs, maxLenError(dns1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		errs = append(errs, regexError(dns1123LabelErrMsg, dns1123LabelFmt, "my-name", "123-abc"))
	}
	return errs
}

// The function was taken as-is from:
// https://github.com/kubernetes/apimachinery/blob/fc49b38c19f02a58ebc476347e622142f19820b9/pkg/util/validation/validation.go
func maxLenError(length int) string {
	return fmt.Sprintf("must be no more than %d characters", length)
}

// The function was taken as-is from: https://github.com/kubernetes/apimachinery/blob/fc49b38c19f02a58ebc476347e622142f19820b9/pkg/util/validation/validation.go
func regexError(msg string, fmt string, examples ...string) string {
	if len(examples) == 0 {
		return msg + " (regex used for validation is '" + fmt + "')"
	}
	msg += " (e.g. "
	for i := range examples {
		if i > 0 {
			msg += " or "
		}
		msg += "'" + examples[i] + "', "
	}
	msg += "regex used for validation is '" + fmt + "')"
	return msg
}
