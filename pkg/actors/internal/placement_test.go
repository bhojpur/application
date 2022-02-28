package internal

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
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	placementv1pb "github.com/bhojpur/application/pkg/api/v1/placement"
	"github.com/bhojpur/application/pkg/placement/hashing"
)

func TestAddDNSResolverPrefix(t *testing.T) {
	testCases := []struct {
		addr          []string
		resolverAdded []string
	}{
		{
			addr:          []string{"placement1:50005", "placement2:50005", "placement3:50005"},
			resolverAdded: []string{"dns:///placement1:50005", "dns:///placement2:50005", "dns:///placement3:50005"},
		}, {
			addr:          []string{"192.168.0.100:50005", "192.168.0.101:50005", "192.168.0.102:50005"},
			resolverAdded: []string{"192.168.0.100:50005", "192.168.0.101:50005", "192.168.0.102:50005"},
		},
	}

	for _, tc := range testCases {
		assert.EqualValues(t, tc.resolverAdded, addDNSResolverPrefix(tc.addr))
	}
}

func TestPlacementStream_RoundRobin(t *testing.T) {
	const testServerCount = 3
	leaderServer := []int32{1, 2}

	address := make([]string, testServerCount)
	testSrv := make([]*testServer, testServerCount)
	cleanup := make([]func(), testServerCount)

	for i := 0; i < testServerCount; i++ {
		address[i], testSrv[i], cleanup[i] = newTestServer()
	}

	appHealthFunc := func() bool {
		return true
	}
	noopTableUpdateFunc := func() {}

	testPlacement := NewActorPlacement(
		address, nil, "testAppID", "127.0.0.1:1000", []string{"actorOne", "actorTwo"},
		appHealthFunc, noopTableUpdateFunc)

	t.Run("found leader placement in a round robin way", func(t *testing.T) {
		// set leader for leaderServer[0]
		testSrv[leaderServer[0]].setLeader(true)

		// act
		testPlacement.Start()
		time.Sleep(statusReportHeartbeatInterval * 3)
		assert.Equal(t, leaderServer[0], testPlacement.serverIndex.Load())
		assert.True(t, testSrv[testPlacement.serverIndex.Load()].recvCount.Load() >= 2)
	})

	t.Run("shutdown leader and find the next leader", func(t *testing.T) {
		// shutdown server
		cleanup[leaderServer[0]]()

		time.Sleep(statusReportHeartbeatInterval)

		// set the second leader
		testSrv[leaderServer[1]].setLeader(true)

		// wait until placement connect to the second leader node
		time.Sleep(statusReportHeartbeatInterval * 3)
		assert.Equal(t, leaderServer[1], testPlacement.serverIndex.Load())
		assert.True(t, testSrv[testPlacement.serverIndex.Load()].recvCount.Load() >= 1)
	})

	// tear down
	testPlacement.Stop()
	time.Sleep(statusReportHeartbeatInterval)
	assert.True(t, testSrv[testPlacement.serverIndex.Load()].isGracefulShutdown.Load())

	for _, fn := range cleanup {
		fn()
	}
}

func TestAppHealthyStatus(t *testing.T) {
	// arrange
	address, testSrv, cleanup := newTestServer()

	// set leader
	testSrv.setLeader(true)

	appHealth := atomic.Bool{}
	appHealth.Store(true)

	appHealthFunc := appHealth.Load
	noopTableUpdateFunc := func() {}
	testPlacement := NewActorPlacement(
		[]string{address}, nil, "testAppID", "127.0.0.1:1000", []string{"actorOne", "actorTwo"},
		appHealthFunc, noopTableUpdateFunc)

	// act
	testPlacement.Start()

	// wait until client sends heartbeat to the test server
	time.Sleep(statusReportHeartbeatInterval * 3)
	oldCount := testSrv.recvCount.Load()
	assert.True(t, oldCount >= 2, "client must send at least twice")

	// Mark app unhealthy
	appHealth.Store(false)
	time.Sleep(statusReportHeartbeatInterval * 2)
	assert.True(t, testSrv.recvCount.Load() <= oldCount+1, "no more +1 heartbeat because app is unhealthy")

	// clean up
	testPlacement.Stop()
	cleanup()
}

func TestOnPlacementOrder(t *testing.T) {
	tableUpdateCount := 0
	appHealthFunc := func() bool { return true }
	tableUpdateFunc := func() { tableUpdateCount++ }
	testPlacement := NewActorPlacement(
		[]string{}, nil,
		"testAppID", "127.0.0.1:1000",
		[]string{"actorOne", "actorTwo"},
		appHealthFunc, tableUpdateFunc)

	t.Run("lock operation", func(t *testing.T) {
		testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{
			Operation: "lock",
		})
		assert.True(t, testPlacement.tableIsBlocked.Load())
	})

	t.Run("update operation", func(t *testing.T) {
		tableVersion := "1"
		tableUpdateCount = 0
		testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{
			Operation: "update",
			Tables: &placementv1pb.PlacementTables{
				Version: tableVersion,
				Entries: map[string]*placementv1pb.PlacementTable{},
			},
		})

		assert.Equal(t, 1, tableUpdateCount)

		// no update with the same table version
		testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{
			Operation: "update",
			Tables: &placementv1pb.PlacementTables{
				Version: tableVersion,
				Entries: map[string]*placementv1pb.PlacementTable{},
			},
		})

		assert.Equal(t, 1, tableUpdateCount)
	})

	t.Run("unlock operation", func(t *testing.T) {
		testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{
			Operation: "unlock",
		})
		assert.False(t, testPlacement.tableIsBlocked.Load())
	})
}

func TestWaitUntilPlacementTableIsReady(t *testing.T) {
	appHealthFunc := func() bool { return true }
	tableUpdateFunc := func() {}
	testPlacement := NewActorPlacement(
		[]string{}, nil,
		"testAppID", "127.0.0.1:1000",
		[]string{"actorOne", "actorTwo"},
		appHealthFunc, tableUpdateFunc)

	testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{Operation: "lock"})

	asserted := atomic.Bool{}
	asserted.Store(false)
	go func() {
		testPlacement.WaitUntilPlacementTableIsReady()
		asserted.Store(true)
	}()

	time.Sleep(50 * time.Millisecond)
	assert.False(t, asserted.Load())

	// unlock
	testPlacement.onPlacementOrder(&placementv1pb.PlacementOrder{Operation: "unlock"})

	// ensure that it is unlocked
	time.Sleep(50 * time.Millisecond)
	assert.True(t, asserted.Load())
}

func TestLookupActor(t *testing.T) {
	appHealthFunc := func() bool { return true }
	tableUpdateFunc := func() {}
	testPlacement := NewActorPlacement(
		[]string{}, nil,
		"testAppID", "127.0.0.1:1000",
		[]string{"actorOne", "actorTwo"},
		appHealthFunc, tableUpdateFunc)

	t.Run("Placementtable is unset", func(t *testing.T) {
		name, appID := testPlacement.LookupActor("actorOne", "test")
		assert.Empty(t, name)
		assert.Empty(t, appID)
	})

	t.Run("found host and appid", func(t *testing.T) {
		const testActorType = "actorOne"
		testPlacement.placementTables = &hashing.ConsistentHashTables{
			Version: "1",
			Entries: map[string]*hashing.Consistent{},
		}

		// set vnode size
		hashing.SetReplicationFactor(10)
		actorOneHashing := hashing.NewConsistentHash()
		actorOneHashing.Add(testPlacement.runtimeHostName, testPlacement.appID, 0)
		testPlacement.placementTables.Entries[testActorType] = actorOneHashing

		// existing actor type
		name, appID := testPlacement.LookupActor(testActorType, "id0")
		assert.Equal(t, testPlacement.runtimeHostName, name)
		assert.Equal(t, testPlacement.appID, appID)

		// non existing actor type
		name, appID = testPlacement.LookupActor("nonExistingActorType", "id0")
		assert.Empty(t, name)
		assert.Empty(t, appID)
	})
}

func TestConcurrentUnblockPlacements(t *testing.T) {
	appHealthFunc := func() bool { return true }
	tableUpdateFunc := func() {}
	testPlacement := NewActorPlacement(
		[]string{}, nil,
		"testAppID", "127.0.0.1:1000",
		[]string{"actorOne", "actorTwo"},
		appHealthFunc, tableUpdateFunc)

	t.Run("concurrent_unlock", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			testPlacement.blockPlacements()
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				testPlacement.unblockPlacements()
				wg.Done()
			}()
			go func() {
				testPlacement.unblockPlacements()
				wg.Done()
			}()
			// Waiting for the goroutines to finish
			wg.Wait()
		}
	})
}

func newTestServer() (string, *testServer, func()) {
	port, _ := freeport.GetFreePort()
	conn := fmt.Sprintf("127.0.0.1:%d", port)
	listener, _ := net.Listen("tcp", conn)
	server := grpc.NewServer()
	srv := &testServer{}
	srv.isGracefulShutdown.Store(false)
	srv.setLeader(false)
	placementv1pb.RegisterPlacementServer(server, srv)

	go func() {
		server.Serve(listener)
	}()

	// wait until test server starts
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		listener.Close()
		server.Stop()
	}

	return conn, srv, cleanup
}

type testServer struct {
	isLeader           atomic.Bool
	lastHost           *placementv1pb.Host
	recvCount          atomic.Int32
	lastTimestamp      time.Time
	recvError          error
	isGracefulShutdown atomic.Bool
}

func (s *testServer) ReportAppStatus(srv placementv1pb.Placement_ReportAppStatusServer) error {
	for {
		if !s.isLeader.Load() {
			return status.Error(codes.FailedPrecondition, "only leader can serve the request")
		}

		req, err := srv.Recv()
		if err == io.EOF {
			s.isGracefulShutdown.Store(true)
			return nil
		} else if err != nil {
			s.recvError = err
			return nil
		}
		s.recvCount.Inc()
		s.lastHost = req
		s.lastTimestamp = time.Now()
	}
}

func (s *testServer) setLeader(leader bool) {
	s.isLeader.Store(leader)
}
