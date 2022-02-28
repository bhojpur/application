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
	"os"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/pkg/errors"
)

const defaultDirPermission = 0755

func ensureDir(dirName string) error {
	info, err := os.Stat(dirName)
	if !os.IsNotExist(err) && !info.Mode().IsDir() {
		return errors.New("file already existed")
	}

	err = os.Mkdir(dirName, defaultDirPermission)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func makeRaftLogCommand(t CommandType, member AppHostMember) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(t))
	err := codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(member)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalMsgPack(in interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, &codec.MsgpackHandle{})
	err := enc.Encode(in)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalMsgPack(in []byte, out interface{}) error {
	dec := codec.NewDecoderBytes(in, &codec.MsgpackHandle{})
	return dec.Decode(out)
}

func raftAddressForID(id string, nodes []PeerInfo) string {
	for _, node := range nodes {
		if node.ID == id {
			return node.Address
		}
	}

	return ""
}
