package api

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
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"

	b64 "encoding/base64"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bhojpur/service/pkg/utils/logger"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	operatorv1pb "github.com/bhojpur/application/pkg/api/v1/operator"
	app_credentials "github.com/bhojpur/application/pkg/credentials"
	componentsapi "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
	configurationapi "github.com/bhojpur/application/pkg/kubernetes/configuration/v1alpha1"
	subscriptionsapi_v2alpha1 "github.com/bhojpur/application/pkg/kubernetes/subscriptions/v2alpha1"
)

const serverPort = 6500

const (
	APIVersionV1alpha1    = "bhojpur.net/v1alpha1"
	APIVersionV2alpha1    = "bhojpur.net/v2alpha1"
	kubernetesSecretStore = "kubernetes"
)

var log = logger.NewLogger("app.operator.api")

// Server runs the Bhojpur Application API server for components and configurations.
type Server interface {
	Run(certChain *app_credentials.CertChain)
	OnComponentUpdated(component *componentsapi.Component)
}

type apiServer struct {
	operatorv1pb.UnimplementedOperatorServer
	Client client.Client
	// notify all Bhojpur Application runtimes
	connLock          sync.Mutex
	allConnUpdateChan map[string]chan *componentsapi.Component
}

// NewAPIServer returns a new API server.
func NewAPIServer(client client.Client) Server {
	return &apiServer{
		Client:            client,
		allConnUpdateChan: make(map[string]chan *componentsapi.Component),
	}
}

// Run starts a new gRPC server.
func (a *apiServer) Run(certChain *app_credentials.CertChain) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", serverPort))
	if err != nil {
		log.Fatal("error starting tcp listener: %s", err)
	}

	opts, err := app_credentials.GetServerOptions(certChain)
	if err != nil {
		log.Fatal("error creating gRPC options: %s", err)
	}
	s := grpc.NewServer(opts...)
	operatorv1pb.RegisterOperatorServer(s, a)

	log.Info("starting gRPC server")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

func (a *apiServer) OnComponentUpdated(component *componentsapi.Component) {
	a.connLock.Lock()
	for _, connUpdateChan := range a.allConnUpdateChan {
		connUpdateChan <- component
	}
	a.connLock.Unlock()
}

// GetConfiguration returns a Bhojpur Application runtime configuration.
func (a *apiServer) GetConfiguration(ctx context.Context, in *operatorv1pb.GetConfigurationRequest) (*operatorv1pb.GetConfigurationResponse, error) {
	key := types.NamespacedName{Namespace: in.Namespace, Name: in.Name}
	var config configurationapi.Configuration
	if err := a.Client.Get(ctx, key, &config); err != nil {
		return nil, errors.Wrap(err, "error getting configuration")
	}
	b, err := json.Marshal(&config)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling configuration")
	}
	return &operatorv1pb.GetConfigurationResponse{
		Configuration: b,
	}, nil
}

// ListComponents returns a list of Bhojpur Application components.
func (a *apiServer) ListComponents(ctx context.Context, in *operatorv1pb.ListComponentsRequest) (*operatorv1pb.ListComponentResponse, error) {
	var components componentsapi.ComponentList
	if err := a.Client.List(ctx, &components, &client.ListOptions{
		Namespace: in.Namespace,
	}); err != nil {
		return nil, errors.Wrap(err, "error getting components")
	}
	resp := &operatorv1pb.ListComponentResponse{
		Components: [][]byte{},
	}
	for i := range components.Items {
		c := components.Items[i] // Make a copy since we will refer to this as a reference in this loop.
		err := processComponentSecrets(&c, in.Namespace, a.Client)
		if err != nil {
			log.Warnf("error processing component %s secrets from pod %s/%s: %s", c.Name, in.Namespace, in.PodName, err)
			return &operatorv1pb.ListComponentResponse{}, err
		}

		b, err := json.Marshal(&c)
		if err != nil {
			log.Warnf("error marshalling component %s from pod %s/%s: %s", c.Name, in.Namespace, in.PodName, err)
			continue
		}
		resp.Components = append(resp.Components, b)
	}
	return resp, nil
}

func processComponentSecrets(component *componentsapi.Component, namespace string, kubeClient client.Client) error {
	for i, m := range component.Spec.Metadata {
		if m.SecretKeyRef.Name != "" && (component.Auth.SecretStore == kubernetesSecretStore || component.Auth.SecretStore == "") {
			var secret corev1.Secret

			err := kubeClient.Get(context.TODO(), types.NamespacedName{
				Name:      m.SecretKeyRef.Name,
				Namespace: namespace,
			}, &secret)
			if err != nil {
				return err
			}

			key := m.SecretKeyRef.Key
			if key == "" {
				key = m.SecretKeyRef.Name
			}

			val, ok := secret.Data[key]
			enc := b64.StdEncoding.EncodeToString(val)
			jsonEnc, err := json.Marshal(enc)
			if err != nil {
				return err
			}

			if ok {
				component.Spec.Metadata[i].Value = componentsapi.DynamicValue{
					JSON: v1.JSON{
						Raw: jsonEnc,
					},
				}
			}
		}
	}

	return nil
}

// ListSubscriptions returns a list of Bhojpur Application pub/sub subscriptions.
func (a *apiServer) ListSubscriptions(ctx context.Context, in *emptypb.Empty) (*operatorv1pb.ListSubscriptionsResponse, error) {
	return a.ListSubscriptionsV2(ctx, &operatorv1pb.ListSubscriptionsRequest{})
}

// ListSubscriptionsV2 returns a list of Bhojpur Application pub/sub subscriptions. Use ListSubscriptionsRequest to expose pod info.
func (a *apiServer) ListSubscriptionsV2(ctx context.Context, in *operatorv1pb.ListSubscriptionsRequest) (*operatorv1pb.ListSubscriptionsResponse, error) {
	resp := &operatorv1pb.ListSubscriptionsResponse{
		Subscriptions: [][]byte{},
	}

	// Only the latest/storage version needs to be returned.
	var subsV2alpha1 subscriptionsapi_v2alpha1.SubscriptionList
	if err := a.Client.List(ctx, &subsV2alpha1); err != nil {
		return nil, errors.Wrap(err, "error getting subscriptions")
	}
	for i := range subsV2alpha1.Items {
		s := subsV2alpha1.Items[i] // Make a copy since we will refer to this as a reference in this loop.
		if s.APIVersion != APIVersionV2alpha1 {
			continue
		}
		b, err := json.Marshal(&s)
		if err != nil {
			log.Warnf("error marshalling subscription for pod %s/%s: %s", in.Namespace, in.PodName, err)
			continue
		}
		resp.Subscriptions = append(resp.Subscriptions, b)
	}

	return resp, nil
}

// ComponentUpdate updates Bhojpur Application runtime sidecars whenever a component in the
// cluster is modified.
func (a *apiServer) ComponentUpdate(in *operatorv1pb.ComponentUpdateRequest, srv operatorv1pb.Operator_ComponentUpdateServer) error {
	log.Info("sidecar connected for component updates")
	key := uuid.New().String()
	a.connLock.Lock()
	a.allConnUpdateChan[key] = make(chan *componentsapi.Component, 1)
	updateChan := a.allConnUpdateChan[key]
	a.connLock.Unlock()
	defer func() {
		a.connLock.Lock()
		delete(a.allConnUpdateChan, key)
		a.connLock.Unlock()
	}()
	chWrapper := initChanGracefully(updateChan)
	updateComponentFunc := func(c *componentsapi.Component) {
		if c.Namespace != in.Namespace {
			return
		}

		err := processComponentSecrets(c, in.Namespace, a.Client)
		if err != nil {
			log.Warnf("error processing component %s secrets from pod %s/%s: %s", c.Name, in.Namespace, in.PodName, err)
			return
		}

		b, err := json.Marshal(&c)
		if err != nil {
			log.Warnf("error serializing component %s (%s) from pod %s/%s: %s", c.GetName(), c.Spec.Type, in.Namespace, in.PodName, err)
			return
		}
		err = srv.Send(&operatorv1pb.ComponentUpdateEvent{
			Component: b,
		})
		if err != nil {
			log.Warnf("error updating sidecar with component %s (%s) from pod %s/%s: %s", c.GetName(), c.Spec.Type, in.Namespace, in.PodName, err)
			if status.Code(err) == codes.Unavailable {
				chWrapper.Close()
			}
			return
		}
		log.Infof("updated sidecar with component %s (%s) from pod %s/%s", c.GetName(), c.Spec.Type, in.Namespace, in.PodName)
	}
	for {
		select {
		case <-srv.Context().Done():
			return nil
		case c, ok := <-updateChan:
			if !ok {
				return nil
			}
			go updateComponentFunc(c)
		}
	}
}

// chanGracefully control channel to close gracefully in multi-goroutines.
type chanGracefully struct {
	ch       chan *componentsapi.Component
	isClosed bool
	sync.Mutex
}

func initChanGracefully(ch chan *componentsapi.Component) (
	c *chanGracefully) {
	return &chanGracefully{
		ch:       ch,
		isClosed: false,
	}
}

// Close chan be closed non-reentrantly.
func (c *chanGracefully) Close() {
	c.Lock()
	if !c.isClosed {
		c.isClosed = true
		close(c.ch)
	}
	c.Unlock()
}
