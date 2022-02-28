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
	"io"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-msgpack/codec"

	"github.com/bhojpur/application/pkg/placement/hashing"
)

// AppHostMember represents Bhojpur Applicaiton runtime actor host member which serve actor types.
type AppHostMember struct {
	// Name is the unique name of Bhojpur Application runtime host.
	Name string
	// AppID is Bhojpur Application runtime app ID.
	AppID string
	// Entities is the list of Actor Types which this Bhojpur Application runtime supports.
	Entities []string

	// UpdatedAt is the last time when this host member info is updated.
	UpdatedAt int64
}

type AppHostMemberStateData struct {
	// Index is the index number of raft log.
	Index uint64
	// Members includes Bhojpur Application runtime hosts.
	Members map[string]*AppHostMember

	// TableGeneration is the generation of hashingTableMap.
	// This is increased whenever hashingTableMap is updated.
	TableGeneration uint64

	// hashingTableMap is the map for storing consistent hashing data
	// per Actor types. This will be generated when log entries are replayed.
	// While snapshotting the state, this member will not be saved. Instead,
	// hashingTableMap will be recovered in snapshot recovery process.
	hashingTableMap map[string]*hashing.Consistent
}

// AppHostMemberState is the state to store Bhojpur Application runtime host and
// consistent hashing tables.
type AppHostMemberState struct {
	lock sync.RWMutex

	data AppHostMemberStateData
}

func newAppHostMemberState() *AppHostMemberState {
	return &AppHostMemberState{
		data: AppHostMemberStateData{
			Index:           0,
			TableGeneration: 0,
			Members:         map[string]*AppHostMember{},
			hashingTableMap: map[string]*hashing.Consistent{},
		},
	}
}

func (s *AppHostMemberState) Index() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.Index
}

func (s *AppHostMemberState) Members() map[string]*AppHostMember {
	s.lock.RLock()
	defer s.lock.RUnlock()

	members := make(map[string]*AppHostMember)
	for k, v := range s.data.Members {
		members[k] = v
	}
	return members
}

func (s *AppHostMemberState) TableGeneration() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.TableGeneration
}

func (s *AppHostMemberState) hashingTableMap() map[string]*hashing.Consistent {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.hashingTableMap
}

func (s *AppHostMemberState) clone() *AppHostMemberState {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newMembers := &AppHostMemberState{
		data: AppHostMemberStateData{
			Index:           s.data.Index,
			TableGeneration: s.data.TableGeneration,
			Members:         map[string]*AppHostMember{},
			hashingTableMap: nil,
		},
	}
	for k, v := range s.data.Members {
		m := &AppHostMember{
			Name:      v.Name,
			AppID:     v.AppID,
			Entities:  make([]string, len(v.Entities)),
			UpdatedAt: v.UpdatedAt,
		}
		copy(m.Entities, v.Entities)
		newMembers.data.Members[k] = m
	}
	return newMembers
}

// caller should holds lock.
func (s *AppHostMemberState) updateHashingTables(host *AppHostMember) {
	for _, e := range host.Entities {
		if _, ok := s.data.hashingTableMap[e]; !ok {
			s.data.hashingTableMap[e] = hashing.NewConsistentHash()
		}

		s.data.hashingTableMap[e].Add(host.Name, host.AppID, 0)
	}
}

// caller should holds lock.
func (s *AppHostMemberState) removeHashingTables(host *AppHostMember) {
	for _, e := range host.Entities {
		if t, ok := s.data.hashingTableMap[e]; ok {
			t.Remove(host.Name)

			// if no dedicated actor service instance for the particular actor type,
			// we must delete consistent hashing table to avoid the memory leak.
			if len(t.Hosts()) == 0 {
				delete(s.data.hashingTableMap, e)
			}
		}
	}
}

// upsertMember upserts member host info to the FSM state and returns true
// if the hashing table update happens.
func (s *AppHostMemberState) upsertMember(host *AppHostMember) bool {
	if !s.isActorHost(host) {
		return false
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if m, ok := s.data.Members[host.Name]; ok {
		// No need to update consistent hashing table if the same Bhojpur Application host member exists
		if m.AppID == host.AppID && m.Name == host.Name && cmp.Equal(m.Entities, host.Entities) {
			m.UpdatedAt = host.UpdatedAt
			return false
		}

		// Remove hashing table because the existing member is invalid
		// and needs to be updated by new member info.
		s.removeHashingTables(m)
	}

	s.data.Members[host.Name] = &AppHostMember{
		Name:      host.Name,
		AppID:     host.AppID,
		UpdatedAt: host.UpdatedAt,
	}

	// Update hashing table only when host reports actor types
	s.data.Members[host.Name].Entities = make([]string, len(host.Entities))
	copy(s.data.Members[host.Name].Entities, host.Entities)

	s.updateHashingTables(s.data.Members[host.Name])

	// Increase hashing table generation version. Runtime will compare the table generation
	// version with its own and then update it if it is new.
	s.data.TableGeneration++

	return true
}

// removeMember removes members from membership and update hashing table and returns true
// if hashing table update happens.
func (s *AppHostMemberState) removeMember(host *AppHostMember) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if m, ok := s.data.Members[host.Name]; ok {
		s.removeHashingTables(m)
		s.data.TableGeneration++
		delete(s.data.Members, host.Name)

		return true
	}

	return false
}

func (s *AppHostMemberState) isActorHost(host *AppHostMember) bool {
	return len(host.Entities) > 0
}

// caller should holds lock.
func (s *AppHostMemberState) restoreHashingTables() {
	if s.data.hashingTableMap == nil {
		s.data.hashingTableMap = map[string]*hashing.Consistent{}
	}

	for _, m := range s.data.Members {
		s.updateHashingTables(m)
	}
}

func (s *AppHostMemberState) restore(r io.Reader) error {
	dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
	var data AppHostMemberStateData
	if err := dec.Decode(&data); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data

	s.restoreHashingTables()
	return nil
}

func (s *AppHostMemberState) persist(w io.Writer) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	b, err := marshalMsgPack(s.data)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
