package actors

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

// OperationType describes a CRUD operation performed against a state store.
type OperationType string

// Upsert is an update or create operation.
const Upsert OperationType = "upsert"

// Delete is a delete operation.
const Delete OperationType = "delete"

// TransactionalRequest describes a set of stateful operations for a given actor that are performed in a transactional manner.
type TransactionalRequest struct {
	Operations []TransactionalOperation `json:"operations"`
	ActorType  string
	ActorID    string
}

// TransactionalOperation is the request object for a state operation participating in a transaction.
type TransactionalOperation struct {
	Operation OperationType `json:"operation"`
	Request   interface{}   `json:"request"`
}

// TransactionalUpsert defines a key/value pair for an upsert operation.
type TransactionalUpsert struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// TransactionalDelete defined a delete operation.
type TransactionalDelete struct {
	Key string `json:"key"`
}
