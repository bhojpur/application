package engine

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
	"strings"
)

// Errors is a struct that used to hold errors array
type Errors struct {
	errors []error
}

// Error get formatted error message
func (errs Errors) Error() string {
	var errors []string
	for _, err := range errs.errors {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, "; ")
}

// AddError add error to Errors struct
func (errs *Errors) AddError(errors ...error) {
	for _, err := range errors {
		if err != nil {
			if e, ok := err.(errorsInterface); ok {
				errs.errors = append(errs.errors, e.GetErrors()...)
			} else {
				errs.errors = append(errs.errors, err)
			}
		}
	}
}

// HasError return has error or not
func (errs Errors) HasError() bool {
	return len(errs.errors) != 0
}

// GetErrors return error array
func (errs Errors) GetErrors() []error {
	return errs.errors
}

type errorsInterface interface {
	GetErrors() []error
}
