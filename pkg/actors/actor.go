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
	"time"

	"go.uber.org/atomic"

	"github.com/pkg/errors"

	diag "github.com/bhojpur/application/pkg/diagnostics"
)

// ErrActorDisposed is the error when runtime tries to hold the lock of the disposed actor.
var ErrActorDisposed error = errors.New("actor is already disposed")

// actor represents single actor object and maintains its turn-based concurrency.
type actor struct {
	// actorType is the type of actor.
	actorType string
	// actorID is the ID of actorType.
	actorID string

	// actorLock is the lock to maintain actor's turn-based concurrency with allowance for reentrancy if configured.
	actorLock ActorLock
	// pendingActorCalls is the number of the current pending actor calls by turn-based concurrency.
	pendingActorCalls atomic.Int32

	// When consistent hashing tables are updated, actor runtime drains actor to rebalance actors
	// across actor hosts after drainOngoingCallTimeout or until all pending actor calls are completed.
	// lastUsedTime is the time when the last actor call holds lock. This is used to calculate
	// the duration of ongoing calls to time out.
	lastUsedTime time.Time

	// disposeLock guards disposed and disposeCh.
	disposeLock *sync.RWMutex
	// disposed is true when actor is already disposed.
	disposed bool
	// disposeCh is the channel to signal when all pending actor calls are completed. This channel
	// is used when runtime drains actor.
	disposeCh chan struct{}

	once sync.Once
}

func newActor(actorType, actorID string, maxReentrancyDepth *int) *actor {
	return &actor{
		actorType:    actorType,
		actorID:      actorID,
		actorLock:    NewActorLock(int32(*maxReentrancyDepth)),
		disposeLock:  &sync.RWMutex{},
		disposeCh:    nil,
		disposed:     false,
		lastUsedTime: time.Now().UTC(),
	}
}

// isBusy returns true when pending actor calls are ongoing.
func (a *actor) isBusy() bool {
	a.disposeLock.RLock()
	disposed := a.disposed
	a.disposeLock.RUnlock()
	return !disposed && a.pendingActorCalls.Load() > 0
}

// channel creates or get new dispose channel. This channel is used for draining the actor.
func (a *actor) channel() chan struct{} {
	a.once.Do(func() {
		a.disposeLock.Lock()
		a.disposeCh = make(chan struct{})
		a.disposeLock.Unlock()
	})

	a.disposeLock.RLock()
	defer a.disposeLock.RUnlock()
	return a.disposeCh
}

// lock holds the lock for turn-based concurrency.
func (a *actor) lock(reentrancyID *string) error {
	pending := a.pendingActorCalls.Inc()
	diag.DefaultMonitoring.ReportActorPendingCalls(a.actorType, pending)

	err := a.actorLock.Lock(reentrancyID)
	if err != nil {
		return err
	}

	a.disposeLock.RLock()
	disposed := a.disposed
	a.disposeLock.RUnlock()
	if disposed {
		a.unlock()
		return ErrActorDisposed
	}
	a.lastUsedTime = time.Now().UTC()
	return nil
}

// unlock releases the lock for turn-based concurrency. If disposeCh is available,
// it will close the channel to notify runtime to dispose actor.
func (a *actor) unlock() {
	pending := a.pendingActorCalls.Dec()
	if pending == 0 {
		func() {
			a.disposeLock.Lock()
			defer a.disposeLock.Unlock()
			if !a.disposed && a.disposeCh != nil {
				a.disposed = true
				close(a.disposeCh)
			}
		}()
	} else if pending < 0 {
		log.Error("BUGBUG: tried to unlock actor before locking actor.")
		return
	}

	a.actorLock.Unlock()
	diag.DefaultMonitoring.ReportActorPendingCalls(a.actorType, pending)
}
