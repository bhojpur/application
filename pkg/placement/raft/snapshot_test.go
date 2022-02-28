package raft

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
	"bytes"
	"testing"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
)

type MockSnapShotSink struct {
	*bytes.Buffer
	cancel bool
}

func (m *MockSnapShotSink) ID() string {
	return "Mock"
}

func (m *MockSnapShotSink) Cancel() error {
	m.cancel = true
	return nil
}

func (m *MockSnapShotSink) Close() error {
	return nil
}

func TestPersist(t *testing.T) {
	// arrange
	fsm := newFSM()
	testMember := AppHostMember{
		Name:     "127.0.0.1:3030",
		AppID:    "fakeAppID",
		Entities: []string{"actorTypeOne", "actorTypeTwo"},
	}
	cmdLog, _ := makeRaftLogCommand(MemberUpsert, testMember)
	raftLog := &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  cmdLog,
	}
	fsm.Apply(raftLog)
	buf := bytes.NewBuffer(nil)
	fakeSink := &MockSnapShotSink{buf, false}

	// act
	snap, err := fsm.Snapshot()
	assert.NoError(t, err)
	snap.Persist(fakeSink)

	// assert
	restoredState := newAppHostMemberState()
	err = restoredState.restore(buf)
	assert.NoError(t, err)

	expectedMember := fsm.State().Members()[testMember.Name]
	restoredMember := restoredState.Members()[testMember.Name]
	assert.Equal(t, fsm.State().Index(), restoredState.Index())
	assert.Equal(t, expectedMember.Name, restoredMember.Name)
	assert.Equal(t, expectedMember.AppID, restoredMember.AppID)
	assert.EqualValues(t, expectedMember.Entities, restoredMember.Entities)
}
