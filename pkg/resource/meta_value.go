package resource

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
	"reflect"

	appsvr "github.com/bhojpur/application/pkg/engine"
)

// MetaValues is slice of MetaValue
type MetaValues struct {
	Values []*MetaValue
}

// Get gets meta value from MetaValues with name
func (mvs MetaValues) Get(name string) *MetaValue {
	for _, mv := range mvs.Values {
		if mv.Name == name {
			return mv
		}
	}

	return nil
}

// MetaValue a struct used to hold information when convert inputs from HTTP
// form, JSON, CSV files and so on to meta values. It will includes field name,
// field value and its configured Meta, if it is a nested resource, will
// includes nested metas in its MetaValues
type MetaValue struct {
	Name       string
	Value      interface{}
	Index      int
	Meta       Metaor
	MetaValues *MetaValues
}

func decodeMetaValuesToField(res Resourcer, field reflect.Value, metaValue *MetaValue, context *appsvr.Context) {
	if field.Kind() == reflect.Struct {
		value := reflect.New(field.Type())
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			field.Set(value.Elem())
		}
	} else if field.Kind() == reflect.Slice {
		if metaValue.Index == 0 {
			field.Set(reflect.Zero(field.Type()))
		}

		var fieldType = field.Type().Elem()
		var isPtr bool
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtr = true
		}

		value := reflect.New(fieldType)
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			if !reflect.DeepEqual(reflect.Zero(fieldType).Interface(), value.Elem().Interface()) {
				if isPtr {
					field.Set(reflect.Append(field, value))
				} else {
					field.Set(reflect.Append(field, value.Elem()))
				}
			}
		}
	}
}
