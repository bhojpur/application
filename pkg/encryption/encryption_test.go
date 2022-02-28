package encryption

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
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
	"github.com/bhojpur/service/pkg/secretstores"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockSecretStore struct {
	secretstores.SecretStore
	primaryKey   string
	secondaryKey string
}

func (m *mockSecretStore) Init(metadata secretstores.Metadata) error {
	if val, ok := metadata.Properties["primaryKey"]; ok {
		m.primaryKey = val
	}

	if val, ok := metadata.Properties["secondaryKey"]; ok {
		m.secondaryKey = val
	}

	return nil
}

func (m *mockSecretStore) GetSecret(req secretstores.GetSecretRequest) (secretstores.GetSecretResponse, error) {
	return secretstores.GetSecretResponse{
		Data: map[string]string{
			"primaryKey":   m.primaryKey,
			"secondaryKey": m.secondaryKey,
		},
	}, nil
}

func (m *mockSecretStore) BulkGetSecret(req secretstores.BulkGetSecretRequest) (secretstores.BulkGetSecretResponse, error) {
	return secretstores.BulkGetSecretResponse{}, nil
}

func TestComponentEncryptionKey(t *testing.T) {
	t.Run("component has a primary and secondary encryption keys", func(t *testing.T) {
		component := v1alpha1.Component{
			ObjectMeta: metav1.ObjectMeta{
				Name: "statestore",
			},
			Spec: v1alpha1.ComponentSpec{
				Metadata: []v1alpha1.MetadataItem{
					{
						Name: primaryEncryptionKey,
						SecretKeyRef: v1alpha1.SecretKeyRef{
							Name: "primaryKey",
						},
					},
					{
						Name: secondaryEncryptionKey,
						SecretKeyRef: v1alpha1.SecretKeyRef{
							Name: "secondaryKey",
						},
					},
				},
			},
		}

		bytes := make([]byte, 32)
		rand.Read(bytes)

		primaryKey := hex.EncodeToString(bytes)

		rand.Read(bytes)

		secondaryKey := hex.EncodeToString(bytes)

		secretStore := &mockSecretStore{}
		secretStore.Init(secretstores.Metadata{
			Properties: map[string]string{
				"primaryKey":   primaryKey,
				"secondaryKey": secondaryKey,
			},
		})

		keys, err := ComponentEncryptionKey(component, secretStore)
		assert.NoError(t, err)
		assert.Equal(t, primaryKey, keys.Primary.Key)
		assert.Equal(t, secondaryKey, keys.Secondary.Key)
	})

	t.Run("keys empty when no secret store is present and no error", func(t *testing.T) {
		component := v1alpha1.Component{
			ObjectMeta: metav1.ObjectMeta{
				Name: "statestore",
			},
			Spec: v1alpha1.ComponentSpec{
				Metadata: []v1alpha1.MetadataItem{
					{
						Name: primaryEncryptionKey,
						SecretKeyRef: v1alpha1.SecretKeyRef{
							Name: "primaryKey",
						},
					},
					{
						Name: secondaryEncryptionKey,
						SecretKeyRef: v1alpha1.SecretKeyRef{
							Name: "secondaryKey",
						},
					},
				},
			},
		}

		keys, err := ComponentEncryptionKey(component, nil)
		assert.Empty(t, keys.Primary.Key)
		assert.Empty(t, keys.Secondary.Key)
		assert.NoError(t, err)
	})

	t.Run("no error when component doesn't have encryption keys", func(t *testing.T) {
		component := v1alpha1.Component{
			ObjectMeta: metav1.ObjectMeta{
				Name: "statestore",
			},
			Spec: v1alpha1.ComponentSpec{
				Metadata: []v1alpha1.MetadataItem{
					{
						Name: "something",
					},
				},
			},
		}

		_, err := ComponentEncryptionKey(component, nil)
		assert.NoError(t, err)
	})
}

func TestTryGetEncryptionKeyFromMetadataItem(t *testing.T) {
	t.Run("no secretRef on valid item", func(t *testing.T) {
		secretStore := &mockSecretStore{}
		secretStore.Init(secretstores.Metadata{
			Properties: map[string]string{
				"primaryKey":   "123",
				"secondaryKey": "456",
			},
		})

		_, err := tryGetEncryptionKeyFromMetadataItem("", v1alpha1.MetadataItem{}, secretStore)
		assert.Error(t, err)
	})
}

func TestCreateCipher(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		gcm, err := createCipher(Key{
			Key: "123",
		}, AES256Algorithm)

		assert.Nil(t, gcm)
		assert.Error(t, err)
	})

	t.Run("valid key", func(t *testing.T) {
		bytes := make([]byte, 32)
		rand.Read(bytes)

		key := hex.EncodeToString(bytes)

		gcm, err := createCipher(Key{
			Key: key,
		}, AES256Algorithm)

		assert.NotNil(t, gcm)
		assert.NoError(t, err)
	})
}
