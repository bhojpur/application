package placement

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
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1pb "github.com/bhojpur/api/pkg/core/v1/placement"
	"github.com/bhojpur/application/pkg/placement/raft"
)

const testStreamSendLatency = 50 * time.Millisecond

var testRaftServer *raft.Server

// TestMain is executed only one time in the entire package to
// start test raft server.
func TestMain(m *testing.M) {
	testRaftServer = raft.New("testnode", true, []raft.PeerInfo{
		{
			ID:      "testnode",
			Address: "127.0.0.1:6060",
		},
	}, "")

	testRaftServer.StartRaft(nil)

	// Wait until test raft node become a leader.
	for range time.Tick(200 * time.Millisecond) {
		if testRaftServer.IsLeader() {
			break
		}
	}

	retVal := m.Run()

	testRaftServer.Shutdown()

	os.Exit(retVal)
}

func newTestPlacementServer(raftServer *raft.Server) (string, *Service, func()) {
	testServer := NewPlacementService(raftServer)

	port, _ := freeport.GetFreePort()
	go func() {
		testServer.Run(strconv.Itoa(port), nil)
	}()

	// Wait until test server starts
	time.Sleep(100 * time.Millisecond)

	cleanUpFn := func() {
		testServer.hasLeadership.Store(false)
		testServer.Shutdown()
	}

	serverAddress := "127.0.0.1:" + strconv.Itoa(port)
	return serverAddress, testServer, cleanUpFn
}

func newTestClient(serverAddress string) (*grpc.ClientConn, v1pb.Placement_ReportAppStatusClient, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	client := v1pb.NewPlacementClient(conn)
	stream, err := client.ReportAppStatus(context.Background())
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, stream, nil
}

func TestMemberRegistration_NoLeadership(t *testing.T) {
	// set up
	serverAddress, testServer, cleanup := newTestPlacementServer(testRaftServer)
	testServer.hasLeadership.Store(false)

	// arrange
	conn, stream, err := newTestClient(serverAddress)
	assert.NoError(t, err)

	host := &v1pb.Host{
		Name:     "127.0.0.1:50102",
		Entities: []string{"DogActor", "CatActor"},
		Id:       "testAppID",
		Load:     1, // Not used yet
		// Port is redundant because Name should include port number
	}

	// act
	stream.Send(host)
	_, err = stream.Recv()
	s, ok := status.FromError(err)

	// assert
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	stream.CloseSend()

	// tear down
	conn.Close()
	cleanup()
}

func TestMemberRegistration_Leadership(t *testing.T) {
	serverAddress, testServer, cleanup := newTestPlacementServer(testRaftServer)
	testServer.hasLeadership.Store(true)

	t.Run("Connect server and disconnect it gracefully", func(t *testing.T) {
		// arrange
		conn, stream, err := newTestClient(serverAddress)
		assert.NoError(t, err)

		host := &v1pb.Host{
			Name:     "127.0.0.1:50102",
			Entities: []string{"DogActor", "CatActor"},
			Id:       "testAppID",
			Load:     1, // Not used yet
			// Port is redundant because Name should include port number
		}

		// act
		stream.Send(host)

		// assert
		select {
		case memberChange := <-testServer.membershipCh:
			assert.Equal(t, raft.MemberUpsert, memberChange.cmdType)
			assert.Equal(t, host.Name, memberChange.host.Name)
			assert.Equal(t, host.Id, memberChange.host.AppID)
			assert.EqualValues(t, host.Entities, memberChange.host.Entities)
			assert.Equal(t, 1, len(testServer.streamConnPool))

		case <-time.After(testStreamSendLatency):
			assert.True(t, false, "no membership change")
		}

		// act
		// Runtime needs to close stream gracefully which will let placement remove runtime host from hashing ring
		// in the next flush time window.
		stream.CloseSend()

		// assert
		select {
		case memberChange := <-testServer.membershipCh:
			assert.Equal(t, raft.MemberRemove, memberChange.cmdType)
			assert.Equal(t, host.Name, memberChange.host.Name)

		case <-time.After(testStreamSendLatency):
			require.True(t, false, "no membership change")
		}

		conn.Close()
	})

	t.Run("Connect server and disconnect it forcefully", func(t *testing.T) {
		// arrange
		conn, stream, err := newTestClient(serverAddress)
		assert.NoError(t, err)

		// act
		host := &v1pb.Host{
			Name:     "127.0.0.1:50103",
			Entities: []string{"DogActor", "CatActor"},
			Id:       "testAppID",
			Load:     1, // Not used yet
			// Port is redundant because Name should include port number
		}
		stream.Send(host)

		// assert
		select {
		case memberChange := <-testServer.membershipCh:
			assert.Equal(t, raft.MemberUpsert, memberChange.cmdType)
			assert.Equal(t, host.Name, memberChange.host.Name)
			assert.Equal(t, host.Id, memberChange.host.AppID)
			assert.EqualValues(t, host.Entities, memberChange.host.Entities)
			assert.Equal(t, 1, len(testServer.streamConnPool))

		case <-time.After(testStreamSendLatency):
			require.True(t, false, "no membership change")
		}

		// act
		// Close tcp connection before closing stream, which simulates the scenario
		// where Bhojpur Application runtime disconnects the connection from placement service unexpectedly.
		conn.Close()

		// assert
		select {
		case <-testServer.membershipCh:
			require.True(t, false, "should not have any member change message because faulty host detector time will clean up")

		case <-time.After(testStreamSendLatency):
			testServer.streamConnPoolLock.RLock()
			streamConnCount := len(testServer.streamConnPool)
			testServer.streamConnPoolLock.RUnlock()
			assert.Equal(t, 0, streamConnCount)
		}
	})

	t.Run("non actor host", func(t *testing.T) {
		// arrange
		conn, stream, err := newTestClient(serverAddress)
		assert.NoError(t, err)

		// act
		host := &v1pb.Host{
			Name:     "127.0.0.1:50104",
			Entities: []string{},
			Id:       "testAppID",
			Load:     1, // Not used yet
			// Port is redundant because Name should include port number
		}
		stream.Send(host)

		// assert
		select {
		case <-testServer.membershipCh:
			require.True(t, false, "should not have any membership change")

		case <-time.After(testStreamSendLatency):
			require.True(t, true)
		}

		// act
		// Close tcp connection before closing stream, which simulates the scenario
		// where Bhojpur Application runtime disconnects the connection from placement service unexpectedly.
		conn.Close()
	})

	cleanup()
}
