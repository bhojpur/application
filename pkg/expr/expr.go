package expr

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
	"encoding/json"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	expr_proto "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const missingVariableMessage = "undeclared reference to '"

type Expr struct {
	expr    string
	program cel.Program
}

func (e *Expr) DecodeString(value string) (err error) {
	var ast *cel.Ast
	var env *cel.Env

	variables := make([]*expr_proto.Decl, 0, 10)
	found := make(map[string]struct{}, 10)

	for {
		env, err = cel.NewEnv(cel.Declarations(variables...))
		if err != nil {
			return err
		}
		var iss *cel.Issues
		ast, iss = env.Compile(value)
		if iss.Err() != nil {
			for _, e := range iss.Errors() {
				if strings.HasPrefix(e.Message, missingVariableMessage) {
					msg := e.Message[len(missingVariableMessage):]
					msg = msg[0:strings.IndexRune(msg, '\'')]
					if _, exists := found[msg]; exists {
						continue
					}
					variables = append(variables, decls.NewVar(msg, decls.Any))
					found[msg] = struct{}{}
				} else {
					return iss.Err()
				}
			}
		} else {
			break
		}
	}
	prg, err := env.Program(ast)
	if err != nil {
		return err
	}

	*e = Expr{
		expr:    value,
		program: prg,
	}

	return nil
}

func (e *Expr) Eval(variables map[string]interface{}) (interface{}, error) {
	out, _, err := e.program.Eval(variables)
	if err != nil {
		return nil, err
	}

	return out.Value(), nil
}

func (e *Expr) Expr() string {
	return e.expr
}

func (e *Expr) String() string {
	return e.expr
}

func (e *Expr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.expr)
}

func (e *Expr) UnmarshalJSON(b []byte) error {
	var code string
	if err := json.Unmarshal(b, &code); err != nil {
		return err
	}

	return e.DecodeString(code)
}
