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

	errsvr "github.com/bhojpur/errors/pkg/validation"
	orm "github.com/bhojpur/orm/pkg/engine"
)

var skipValidations = "validations:skip_validations"

func validate(scope *orm.Scope) {
	if _, ok := scope.Get("orm:update_column"); !ok {
		if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
			if !scope.HasError() {
				scope.CallMethod("Validate")
				if scope.Value != nil {
					resource := scope.IndirectValue().Interface()
					_, validatorErrors := errsvr.ValidateStruct(resource)
					if validatorErrors != nil {
						if errors, ok := validatorErrors.(errsvr.Errors); ok {
							for _, err := range flatValidatorErrors(errors) {
								scope.DB().AddError(formattedError(err, resource))
							}
						} else {
							scope.DB().AddError(validatorErrors)
						}
					}
				}
			}
		}
	}
}

func flatValidatorErrors(validatorErrors errsvr.Errors) []errsvr.Error {
	resultErrors := []errsvr.Error{}
	for _, validatorError := range validatorErrors.Errors() {
		if errors, ok := validatorError.(errsvr.Errors); ok {
			for _, e := range errors {
				resultErrors = append(resultErrors, e.(errsvr.Error))
			}
		}
		if e, ok := validatorError.(errsvr.Error); ok {
			resultErrors = append(resultErrors, e)
		}
	}
	return resultErrors
}

func formattedError(err errsvr.Error, resource interface{}) error {
	message := err.Error()
	attrName := err.Name
	if strings.Index(message, "non zero value required") >= 0 {
		message = fmt.Sprintf("%v can't be blank", attrName)
	} else if strings.Index(message, "as length") >= 0 {
		reg, _ := regexp.Compile(`\(([0-9]+)\|([0-9]+)\)`)
		submatch := reg.FindSubmatch([]byte(err.Error()))
		message = fmt.Sprintf("%v is the wrong length (should be %v~%v characters)", attrName, string(submatch[1]), string(submatch[2]))
	} else if strings.Index(message, "as numeric") >= 0 {
		message = fmt.Sprintf("%v is not a number", attrName)
	} else if strings.Index(message, "as email") >= 0 {
		message = fmt.Sprintf("%v is not a valid email address", attrName)
	}
	return NewError(resource, attrName, message)

}

// RegisterCallbacks register callback into Bhojpur ORM DB
func RegisterCallbacks(db *orm.DB) {
	callback := db.Callback()
	if callback.Create().Get("validations:validate") == nil {
		callback.Create().Before("orm:before_create").Register("validations:validate", validate)
	}
	if callback.Update().Get("validations:validate") == nil {
		callback.Update().Before("orm:before_update").Register("validations:validate", validate)
	}
}
