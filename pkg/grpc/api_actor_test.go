package grpc

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
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/bhojpur/application/pkg/actors"
	runtimev1pb "github.com/bhojpur/api/pkg/core/v1/runtime"
	appt "github.com/bhojpur/application/pkg/testing"
)

func TestRegisterActorReminder(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.RegisterActorReminder(context.TODO(), &runtimev1pb.RegisterActorReminderRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestUnregisterActorTimer(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.UnregisterActorTimer(context.TODO(), &runtimev1pb.UnregisterActorTimerRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestRegisterActorTimer(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.RegisterActorTimer(context.TODO(), &runtimev1pb.RegisterActorTimerRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestGetActorState(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.GetActorState(context.TODO(), &runtimev1pb.GetActorStateRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("Get actor state - OK", func(t *testing.T) {
		data := []byte("{ \"data\": 123 }")
		mockActors := new(appt.MockActors)
		mockActors.On("GetState", &actors.GetStateRequest{
			ActorID:   "fakeActorID",
			ActorType: "fakeActorType",
			Key:       "key1",
		}).Return(&actors.StateResponse{
			Data: data,
		}, nil)

		mockActors.On("IsActorHosted", &actors.ActorHostedRequest{
			ActorID:   "fakeActorID",
			ActorType: "fakeActorType",
		}).Return(true)

		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id:    "fakeAPI",
			actor: mockActors,
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)

		// act
		res, err := client.GetActorState(context.TODO(), &runtimev1pb.GetActorStateRequest{
			ActorId:   "fakeActorID",
			ActorType: "fakeActorType",
			Key:       "key1",
		})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, data, res.Data)
		mockActors.AssertNumberOfCalls(t, "GetState", 1)
	})
}

func TestExecuteActorStateTransaction(t *testing.T) {
	port, _ := freeport.GetFreePort()

	t.Run("actors not initialized", func(t *testing.T) {
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.ExecuteActorStateTransaction(context.TODO(), &runtimev1pb.ExecuteActorStateTransactionRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("Save actor state - Upsert and Delete OK", func(t *testing.T) {
		data := []byte("{ \"data\": 123 }")
		mockActors := new(appt.MockActors)
		mockActors.On("TransactionalStateOperation", &actors.TransactionalRequest{
			ActorID:   "fakeActorID",
			ActorType: "fakeActorType",
			Operations: []actors.TransactionalOperation{
				{
					Operation: "upsert",
					Request: map[string]interface{}{
						"key":   "key1",
						"value": data,
					},
				},
				{
					Operation: "delete",
					Request: map[string]interface{}{
						"key": "key2",
					},
				},
			},
		}).Return(nil)

		mockActors.On("IsActorHosted", &actors.ActorHostedRequest{
			ActorID:   "fakeActorID",
			ActorType: "fakeActorType",
		}).Return(true)

		server := startAppAPIServer(port, &api{
			id:    "fakeAPI",
			actor: mockActors,
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)

		// act
		res, err := client.ExecuteActorStateTransaction(context.TODO(),
			&runtimev1pb.ExecuteActorStateTransactionRequest{
				ActorId:   "fakeActorID",
				ActorType: "fakeActorType",
				Operations: []*runtimev1pb.TransactionalActorStateOperation{
					{
						OperationType: "upsert",
						Key:           "key1",
						Value:         &anypb.Any{Value: data},
					},
					{
						OperationType: "delete",
						Key:           "key2",
					},
				},
			})

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, res)
		mockActors.AssertNumberOfCalls(t, "TransactionalStateOperation", 1)
	})
}

func TestUnregisterActorReminder(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.UnregisterActorReminder(context.TODO(), &runtimev1pb.UnregisterActorReminderRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestInvokeActor(t *testing.T) {
	t.Run("actors not initialized", func(t *testing.T) {
		port, _ := freeport.GetFreePort()
		server := startAppAPIServer(port, &api{
			id: "fakeAPI",
		}, "")
		defer server.Stop()

		clientConn := createTestClient(port)
		defer clientConn.Close()

		client := runtimev1pb.NewApplicationClient(clientConn)
		_, err := client.InvokeActor(context.TODO(), &runtimev1pb.InvokeActorRequest{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
