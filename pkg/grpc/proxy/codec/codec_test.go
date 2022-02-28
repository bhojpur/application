package codec

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

	_ "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	pb "github.com/trusch/grpc-proxy/testservice"
	"google.golang.org/grpc/encoding"
)

func TestCodec_ReadYourWrites(t *testing.T) {
	framePtr := &Frame{}
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	Register()
	codec := encoding.GetCodec((&Proxy{}).Name())
	require.NotNil(t, codec, "codec must be registered")
	require.NoError(t, codec.Unmarshal(data, framePtr), "unmarshalling must go ok")
	out, err := codec.Marshal(framePtr)
	require.NoError(t, err, "no marshal error")
	require.Equal(t, data, out, "output and data must be the same")

	// reuse
	require.NoError(t, codec.Unmarshal([]byte{0x55}, framePtr), "unmarshalling must go ok")
	out, err = codec.Marshal(framePtr)
	require.NoError(t, err, "no marshal error")
	require.Equal(t, []byte{0x55}, out, "output and data must be the same")
}

func TestProtoCodec_ReadYourWrites(t *testing.T) {
	p1 := &pb.PingRequest{
		Value: "test-ping",
	}
	proxyCd := encoding.GetCodec((&Proxy{}).Name())

	require.NotNil(t, proxyCd, "proxy codec must not be nil")

	out1p1, err := proxyCd.Marshal(p1)
	require.NoError(t, err, "marshalling must go ok")
	out2p1, err := proxyCd.Marshal(p1)
	require.NoError(t, err, "marshalling must go ok")

	p2 := &pb.PingRequest{}
	err = proxyCd.Unmarshal(out1p1, p2)
	require.NoError(t, err, "unmarshalling must go ok")
	err = proxyCd.Unmarshal(out2p1, p2)
	require.NoError(t, err, "unmarshalling must go ok")

	require.Equal(t, *p1, *p2)
}
