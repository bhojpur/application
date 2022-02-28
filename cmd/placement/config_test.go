package main

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

	"github.com/bhojpur/application/pkg/placement/raft"
)

func TestParsePeersFromFlag(t *testing.T) {
	peerAddressTests := []struct {
		in  string
		out []raft.PeerInfo
	}{
		{
			"node0=127.0.0.1:3030",
			[]raft.PeerInfo{
				{ID: "node0", Address: "127.0.0.1:3030"},
			},
		}, {
			"node0=127.0.0.1:3030,node1=127.0.0.1:3031,node2=127.0.0.1:3032",
			[]raft.PeerInfo{
				{ID: "node0", Address: "127.0.0.1:3030"},
				{ID: "node1", Address: "127.0.0.1:3031"},
				{ID: "node2", Address: "127.0.0.1:3032"},
			},
		}, {
			"127.0.0.1:3030,node1=127.0.0.1:3031,node2=127.0.0.1:3032",
			[]raft.PeerInfo{
				{ID: "node1", Address: "127.0.0.1:3031"},
				{ID: "node2", Address: "127.0.0.1:3032"},
			},
		},
	}

	for _, tt := range peerAddressTests {
		t.Run("parse peers from cmd flag: "+tt.in, func(t *testing.T) {
			peerInfo := parsePeersFromFlag(tt.in)
			assert.EqualValues(t, tt.out, peerInfo)
		})
	}
}
