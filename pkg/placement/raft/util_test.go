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
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureDir(t *testing.T) {
	testDir := "_testDir"
	t.Run("create dir successfully", func(t *testing.T) {
		err := ensureDir(testDir)
		assert.NoError(t, err)
		err = os.Remove(testDir)
		assert.NoError(t, err)
	})

	t.Run("ensure the existing directory", func(t *testing.T) {
		err := os.Mkdir(testDir, 0700)
		assert.NoError(t, err)
		err = ensureDir(testDir)
		assert.NoError(t, err)
		err = os.Remove(testDir)
		assert.NoError(t, err)
	})

	t.Run("fails to create dir", func(t *testing.T) {
		file, err := os.Create(testDir)
		assert.NoError(t, err)
		log.Printf("%v", file)
		file.Close()
		err = ensureDir(testDir)
		assert.Error(t, err)
		err = os.Remove(testDir)
		assert.NoError(t, err)
	})
}

func TestRaftAddressForID(t *testing.T) {
	raftAddressTests := []struct {
		in  []PeerInfo
		id  string
		out string
	}{
		{
			[]PeerInfo{
				{ID: "node0", Address: "127.0.0.1:3030"},
				{ID: "node1", Address: "127.0.0.1:3031"},
			},
			"node0",
			"127.0.0.1:3030",
		}, {
			[]PeerInfo{
				{ID: "node0", Address: "127.0.0.1:3030"},
			},
			"node1",
			"",
		},
	}

	for _, tt := range raftAddressTests {
		t.Run(fmt.Sprintf("find %s from %v", tt.id, tt.in), func(t *testing.T) {
			assert.Equal(t, tt.out, raftAddressForID(tt.id, tt.in))
		})
	}
}

func TestMarshalAndUnmarshalMsgpack(t *testing.T) {
	type testStruct struct {
		Name            string
		StringArrayList []string
		notSerialized   map[string]string
	}

	testObject := testStruct{
		Name:            "namevalue",
		StringArrayList: []string{"value1", "value2"},
		notSerialized: map[string]string{
			"key": "value",
		},
	}

	encoded, err := marshalMsgPack(testObject)
	assert.NoError(t, err)

	var decoded testStruct
	err = unmarshalMsgPack(encoded, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, testObject.Name, decoded.Name)
	assert.Equal(t, testObject.StringArrayList, decoded.StringArrayList)
	assert.Nil(t, decoded.notSerialized)
}

func TestMakeRaftLogCommand(t *testing.T) {
	// arrange
	testMember := AppHostMember{
		Name:     "127.0.0.1:3030",
		AppID:    "fakeAppID",
		Entities: []string{"actorTypeOne", "actorTypeTwo"},
	}

	// act
	cmdLog, _ := makeRaftLogCommand(MemberUpsert, testMember)

	// assert
	assert.Equal(t, uint8(MemberUpsert), cmdLog[0])
	unmarshaled := AppHostMember{}
	unmarshalMsgPack(cmdLog[1:], &unmarshaled)
	assert.EqualValues(t, testMember, unmarshaled)
}
