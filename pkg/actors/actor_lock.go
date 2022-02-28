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

import (
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

var ErrMaxStackDepthExceeded error = errors.New("Maximum stack depth exceeded")

type ActorLock struct {
	methodLock    *sync.Mutex
	requestLock   *sync.Mutex
	activeRequest *string
	stackDepth    *atomic.Int32
	maxStackDepth int32
}

func NewActorLock(maxStackDepth int32) ActorLock {
	return ActorLock{
		methodLock:    &sync.Mutex{},
		requestLock:   &sync.Mutex{},
		activeRequest: nil,
		stackDepth:    atomic.NewInt32(int32(0)),
		maxStackDepth: maxStackDepth,
	}
}

func (a *ActorLock) Lock(requestID *string) error {
	currentRequest := a.getCurrentID()

	if a.stackDepth.Load() == a.maxStackDepth {
		return ErrMaxStackDepthExceeded
	}

	if currentRequest == nil || *currentRequest != *requestID {
		a.methodLock.Lock()
		a.setCurrentID(requestID)
		a.stackDepth.Inc()
	} else {
		a.stackDepth.Inc()
	}

	return nil
}

func (a *ActorLock) Unlock() {
	a.stackDepth.Dec()
	if a.stackDepth.Load() == 0 {
		a.clearCurrentID()
		a.methodLock.Unlock()
	}
}

func (a *ActorLock) getCurrentID() *string {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	return a.activeRequest
}

func (a *ActorLock) setCurrentID(id *string) {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	a.activeRequest = id
}

func (a *ActorLock) clearCurrentID() {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	a.activeRequest = nil
}
