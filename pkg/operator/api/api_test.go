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
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	operatorv1pb "github.com/bhojpur/application/pkg/api/v1/operator"
	"github.com/bhojpur/application/pkg/client/clientset/versioned/scheme"
	componentsapi "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
)

type mockComponentUpdateServer struct {
	grpc.ServerStream
	Calls int
}

func (m *mockComponentUpdateServer) Send(*operatorv1pb.ComponentUpdateEvent) error {
	m.Calls++
	return nil
}

func (m *mockComponentUpdateServer) Context() context.Context {
	return context.TODO()
}

func TestProcessComponentSecrets(t *testing.T) {
	t.Run("secret ref exists, not kubernetes secret store, no error", func(t *testing.T) {
		c := componentsapi.Component{
			Spec: componentsapi.ComponentSpec{
				Metadata: []componentsapi.MetadataItem{
					{
						Name: "test1",
						SecretKeyRef: componentsapi.SecretKeyRef{
							Name: "secret1",
							Key:  "key1",
						},
					},
				},
			},
			Auth: componentsapi.Auth{
				SecretStore: "secretstore",
			},
		}

		err := processComponentSecrets(&c, "default", nil)
		assert.NoError(t, err)
	})

	t.Run("secret ref exists, kubernetes secret store, secret extracted", func(t *testing.T) {
		c := componentsapi.Component{
			Spec: componentsapi.ComponentSpec{
				Metadata: []componentsapi.MetadataItem{
					{
						Name: "test1",
						SecretKeyRef: componentsapi.SecretKeyRef{
							Name: "secret1",
							Key:  "key1",
						},
					},
				},
			},
			Auth: componentsapi.Auth{
				SecretStore: kubernetesSecretStore,
			},
		}

		s := runtime.NewScheme()
		err := scheme.AddToScheme(s)
		assert.NoError(t, err)

		err = corev1.AddToScheme(s)
		assert.NoError(t, err)

		client := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"key1": []byte("value1"),
				},
			}).
			Build()

		err = processComponentSecrets(&c, "default", client)
		assert.NoError(t, err)

		enc := base64.StdEncoding.EncodeToString([]byte("value1"))
		jsonEnc, _ := json.Marshal(enc)

		assert.Equal(t, jsonEnc, c.Spec.Metadata[0].Value.Raw)
	})

	t.Run("secret ref exists, default kubernetes secret store, secret extracted", func(t *testing.T) {
		c := componentsapi.Component{
			Spec: componentsapi.ComponentSpec{
				Metadata: []componentsapi.MetadataItem{
					{
						Name: "test1",
						SecretKeyRef: componentsapi.SecretKeyRef{
							Name: "secret1",
							Key:  "key1",
						},
					},
				},
			},
			Auth: componentsapi.Auth{
				SecretStore: "",
			},
		}

		s := runtime.NewScheme()
		err := scheme.AddToScheme(s)
		assert.NoError(t, err)

		err = corev1.AddToScheme(s)
		assert.NoError(t, err)

		client := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"key1": []byte("value1"),
				},
			}).
			Build()

		err = processComponentSecrets(&c, "default", client)
		assert.NoError(t, err)

		enc := base64.StdEncoding.EncodeToString([]byte("value1"))
		jsonEnc, _ := json.Marshal(enc)

		assert.Equal(t, jsonEnc, c.Spec.Metadata[0].Value.Raw)
	})
}

func TestChanGracefullyClose(t *testing.T) {
	t.Run("close updateChan", func(t *testing.T) {
		ch := make(chan *componentsapi.Component)
		instance := initChanGracefully(ch)
		instance.Close()
		assert.Equal(t, true, instance.isClosed)
	})
}

func TestComponentUpdate(t *testing.T) {
	t.Run("skip sidecar update if namespace doesn't match", func(t *testing.T) {
		c := componentsapi.Component{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
			},
			Spec: componentsapi.ComponentSpec{},
		}

		s := runtime.NewScheme()
		err := scheme.AddToScheme(s)
		assert.NoError(t, err)

		err = corev1.AddToScheme(s)
		assert.NoError(t, err)

		client := fake.NewClientBuilder().
			WithScheme(s).Build()

		mockSidecar := &mockComponentUpdateServer{}
		api := NewAPIServer(client).(*apiServer)

		go func() {
			// Send a component update, give sidecar time to register
			time.Sleep(time.Millisecond * 500)

			for _, connUpdateChan := range api.allConnUpdateChan {
				connUpdateChan <- &c

				// Give sidecar time to register update
				time.Sleep(time.Millisecond * 500)
				close(connUpdateChan)
			}
		}()

		// Start sidecar update loop
		api.ComponentUpdate(&operatorv1pb.ComponentUpdateRequest{
			Namespace: "ns2",
		}, mockSidecar)

		assert.Zero(t, mockSidecar.Calls)
	})

	t.Run("sidecar is updated when component namespace is a match", func(t *testing.T) {
		c := componentsapi.Component{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
			},
			Spec: componentsapi.ComponentSpec{},
		}

		s := runtime.NewScheme()
		err := scheme.AddToScheme(s)
		assert.NoError(t, err)

		err = corev1.AddToScheme(s)
		assert.NoError(t, err)

		client := fake.NewClientBuilder().
			WithScheme(s).Build()

		mockSidecar := &mockComponentUpdateServer{}
		api := NewAPIServer(client).(*apiServer)

		go func() {
			// Send a component update, give sidecar time to register
			time.Sleep(time.Millisecond * 500)

			for _, connUpdateChan := range api.allConnUpdateChan {
				connUpdateChan <- &c

				// Give sidecar time to register update
				time.Sleep(time.Millisecond * 500)
				close(connUpdateChan)
			}
		}()

		// Start sidecar update loop
		api.ComponentUpdate(&operatorv1pb.ComponentUpdateRequest{
			Namespace: "ns1",
		}, mockSidecar)

		assert.Equal(t, 1, mockSidecar.Calls)
	})
}
