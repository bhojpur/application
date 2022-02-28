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
	"net/http"

	"google.golang.org/grpc"

	"github.com/bhojpur/application/pkg/config"
	v1 "github.com/bhojpur/application/pkg/messaging/v1"
)

var endpoints = map[string][]string{
	"invoke.v1": {
		"/v1.runtime.Application/InvokeService",
	},
	"state.v1": {
		"/v1.runtime.Application/GetState",
		"/v1.runtime.Application/GetBulkState",
		"/v1.runtime.Application/SaveState",
		"/v1.runtime.Application/QueryState",
		"/v1.runtime.application/DeleteState",
		"/v1.runtime.Application/DeleteBulkState",
		"/v1.runtime.Application/ExecuteStateTransaction",
	},
	"publish.v1": {
		"/v1.runtime.Application/PublishEvent",
	},
	"bindings.v1": {
		"/v1.runtime.Application/InvokeBinding",
	},
	"secrets.v1": {
		"/v1.runtime.Application/GetSecret",
		"/v1.runtime.Application/GetBulkSecret",
	},
	"actors.v1": {
		"/v1.runtime.Application/RegisterActorTimer",
		"/v1.runtime.Application/UnregisterActorTimer",
		"/v1.runtime.Application/RegisterActorReminder",
		"/v1.runtime.Application/UnregisterActorReminder",
		"/v1.runtime.Application/RenameActorReminder",
		"/v1.runtime.Application/GetActorState",
		"/v1.runtime.Application/ExecuteActorStateTransaction",
		"/v1.runtime.Application/InvokeActor",
	},
	"metadata.v1": {
		"/v1.runtime.Application/GetMetadata",
		"/v1.runtime.Application/SetMetadata",
	},
	"shutdown.v1": {
		"/v1.runtime.Application/Shutdown",
	},
}

const protocol = "grpc"

func setAPIEndpointsMiddlewareUnary(rules []config.APIAccessRule) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var grpcRules []config.APIAccessRule

		for _, rule := range rules {
			if rule.Protocol == protocol {
				grpcRules = append(grpcRules, rule)
			}
		}

		if len(grpcRules) == 0 {
			return handler(ctx, req)
		}

		for _, rule := range grpcRules {
			if list, ok := endpoints[rule.Name+"."+rule.Version]; ok {
				for _, method := range list {
					if method == info.FullMethod {
						return handler(ctx, req)
					}
				}
			}
		}

		err := v1.ErrorFromHTTPResponseCode(http.StatusNotImplemented, "requested Bhojpur Application runtime endpoint is not available")
		return nil, err
	}
}
