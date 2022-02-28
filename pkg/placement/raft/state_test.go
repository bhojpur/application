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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppHostMemberState(t *testing.T) {
	// act
	s := newAppHostMemberState()

	// assert
	assert.Equal(t, uint64(0), s.Index())
	assert.Equal(t, 0, len(s.Members()))
	assert.Equal(t, 0, len(s.hashingTableMap()))
}

func TestClone(t *testing.T) {
	// arrange
	s := newAppHostMemberState()
	s.upsertMember(&AppHostMember{
		Name:     "127.0.0.1:8080",
		AppID:    "FakeID",
		Entities: []string{"actorTypeOne", "actorTypeTwo"},
	})

	// act
	newState := s.clone()

	// assert
	assert.NotSame(t, s, newState)
	assert.Nil(t, newState.hashingTableMap())
	assert.Equal(t, s.Index(), newState.Index())
	assert.EqualValues(t, s.Members(), newState.Members())
}

func TestUpsertMember(t *testing.T) {
	// arrange
	s := newAppHostMemberState()

	t.Run("add new actor member", func(t *testing.T) {
		// act
		updated := s.upsertMember(&AppHostMember{
			Name:      "127.0.0.1:8080",
			AppID:     "FakeID",
			Entities:  []string{"actorTypeOne", "actorTypeTwo"},
			UpdatedAt: 1,
		})

		// assert
		assert.Equal(t, 1, len(s.Members()))
		assert.Equal(t, 2, len(s.hashingTableMap()))
		assert.True(t, updated)
	})

	t.Run("no hashing table update required", func(t *testing.T) {
		// act
		updated := s.upsertMember(&AppHostMember{
			Name:      "127.0.0.1:8081",
			AppID:     "FakeID_2",
			Entities:  []string{"actorTypeOne", "actorTypeTwo"},
			UpdatedAt: 1,
		})

		// assert
		assert.Equal(t, 2, len(s.Members()))
		assert.Equal(t, 2, len(s.hashingTableMap()))
		assert.True(t, updated)

		// act
		updated = s.upsertMember(&AppHostMember{
			Name:      "127.0.0.1:8081",
			AppID:     "FakeID_2",
			Entities:  []string{"actorTypeOne", "actorTypeTwo"},
			UpdatedAt: 2,
		})

		// assert
		assert.False(t, updated)
	})

	t.Run("non actor host", func(t *testing.T) {
		testMember := &AppHostMember{
			Name:      "127.0.0.1:8080",
			AppID:     "FakeID",
			Entities:  []string{},
			UpdatedAt: 100,
		}

		// act
		updated := s.upsertMember(testMember)
		// assert
		assert.False(t, updated)
	})

	t.Run("update existing actor member", func(t *testing.T) {
		testMember := &AppHostMember{
			Name:      "127.0.0.1:8080",
			AppID:     "FakeID",
			Entities:  []string{"actorTypeThree"},
			UpdatedAt: 100,
		}

		// act
		//
		// this tries to update the existing actor members.
		// it will delete empty consistent hashing table.
		updated := s.upsertMember(testMember)

		// assert
		assert.Equal(t, 2, len(s.Members()))
		assert.True(t, updated)
		assert.Equal(t, 1, len(s.Members()[testMember.Name].Entities))
		assert.Equal(t, 3, len(s.hashingTableMap()), "this doesn't delete empty consistent hashing table")
	})
}

func TestRemoveMember(t *testing.T) {
	// arrange
	s := newAppHostMemberState()

	t.Run("remove member and clean up consistent hashing table", func(t *testing.T) {
		// act
		updated := s.upsertMember(&AppHostMember{
			Name:     "127.0.0.1:8080",
			AppID:    "FakeID",
			Entities: []string{"actorTypeOne", "actorTypeTwo"},
		})

		// assert
		assert.Equal(t, 1, len(s.Members()))
		assert.True(t, updated)
		assert.Equal(t, 2, len(s.hashingTableMap()))

		// act
		updated = s.removeMember(&AppHostMember{
			Name: "127.0.0.1:8080",
		})

		// assert
		assert.Equal(t, 0, len(s.Members()))
		assert.True(t, updated)
		assert.Equal(t, 0, len(s.hashingTableMap()))
	})

	t.Run("no table update required", func(t *testing.T) {
		// act
		updated := s.removeMember(&AppHostMember{
			Name: "127.0.0.1:8080",
		})

		// assert
		assert.Equal(t, 0, len(s.Members()))
		assert.False(t, updated)
		assert.Equal(t, 0, len(s.hashingTableMap()))
	})
}

func TestUpdateHashingTable(t *testing.T) {
	// each subtest has dependency on the state

	// arrange
	s := newAppHostMemberState()

	t.Run("add new hashing table per actor types", func(t *testing.T) {
		testMember := &AppHostMember{
			Name:     "127.0.0.1:8080",
			AppID:    "FakeID",
			Entities: []string{"actorTypeOne", "actorTypeTwo"},
		}

		// act
		s.updateHashingTables(testMember)

		assert.Equal(t, 2, len(s.hashingTableMap()))
		for _, ent := range testMember.Entities {
			assert.NotNil(t, s.hashingTableMap()[ent])
		}
	})

	t.Run("update new hashing table per actor types", func(t *testing.T) {
		testMember := &AppHostMember{
			Name:     "127.0.0.1:8080",
			AppID:    "FakeID",
			Entities: []string{"actorTypeOne", "actorTypeTwo", "actorTypeThree"},
		}

		// act
		s.updateHashingTables(testMember)

		assert.Equal(t, 3, len(s.hashingTableMap()))
		for _, ent := range testMember.Entities {
			assert.NotNil(t, s.hashingTableMap()[ent])
		}
	})
}

func TestRemoveHashingTable(t *testing.T) {
	// each subtest has dependency on the state

	// arrange
	testMember := &AppHostMember{
		Name:     "fakeName",
		AppID:    "fakeID",
		Entities: []string{"actorTypeOne", "actorTypeTwo"},
	}

	testcases := []struct {
		name       string
		totalTable int
	}{
		{"127.0.0.1:8080", 2},
		{"127.0.0.1:8081", 0},
	}

	s := newAppHostMemberState()
	for _, tc := range testcases {
		testMember.Name = tc.name
		s.updateHashingTables(testMember)
	}

	// act
	for _, tc := range testcases {
		t.Run("remove host "+tc.name, func(t *testing.T) {
			testMember.Name = tc.name
			s.removeHashingTables(testMember)

			assert.Equal(t, tc.totalTable, len(s.hashingTableMap()))
		})
	}
}

func TestRestoreHashingTables(t *testing.T) {
	// arrange
	testnames := []string{
		"127.0.0.1:8080",
		"127.0.0.1:8081",
	}

	s := &AppHostMemberState{
		data: AppHostMemberStateData{
			Index:           0,
			Members:         map[string]*AppHostMember{},
			hashingTableMap: nil,
		},
	}
	for _, tn := range testnames {
		s.lock.Lock()
		s.data.Members[tn] = &AppHostMember{
			Name:     tn,
			AppID:    "fakeID",
			Entities: []string{"actorTypeOne", "actorTypeTwo"},
		}
		s.lock.Unlock()
	}
	assert.Equal(t, 0, len(s.hashingTableMap()))

	// act
	s.restoreHashingTables()

	// assert
	assert.Equal(t, 2, len(s.hashingTableMap()))
}
