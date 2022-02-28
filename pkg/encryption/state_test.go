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
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddEncryptedStateStore(t *testing.T) {
	t.Run("state store doesn't exist", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		r := AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: Key{
				Name: "primary",
				Key:  "123",
			},
		})
		assert.True(t, r)
		assert.Equal(t, "primary", encryptedStateStores["test"].Primary.Name)
	})

	t.Run("state store exists", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		r := AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: Key{
				Name: "primary",
				Key:  "123",
			},
		})

		assert.True(t, r)
		assert.Equal(t, "primary", encryptedStateStores["test"].Primary.Name)

		r = AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: Key{
				Name: "primary",
				Key:  "123",
			},
		})

		assert.False(t, r)
	})
}

func TestTryEncryptValue(t *testing.T) {
	t.Run("state store without keys", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		ok := EncryptedStateStore("test")
		assert.False(t, ok)
	})

	t.Run("state store with AES256 primary key, value encrypted and decrypted successfully", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}

		bytes := make([]byte, 32)
		rand.Read(bytes)

		key := hex.EncodeToString(bytes)

		pr := Key{
			Name: "primary",
			Key:  key,
		}

		gcm, _ := createCipher(pr, AES256Algorithm)
		pr.gcm = gcm

		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: pr,
		})

		v := []byte("hello")
		r, err := TryEncryptValue("test", v)

		assert.NoError(t, err)
		assert.NotEqual(t, v, r)

		dr, err := TryDecryptValue("test", r)
		assert.NoError(t, err)
		assert.Equal(t, v, dr)
	})

	t.Run("state store with AES256 secondary key, value encrypted and decrypted successfully", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}

		bytes := make([]byte, 32)
		rand.Read(bytes)

		primaryKey := hex.EncodeToString(bytes)

		pr := Key{
			Name: "primary",
			Key:  primaryKey,
		}

		gcm, _ := createCipher(pr, AES256Algorithm)
		pr.gcm = gcm

		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: pr,
		})

		v := []byte("hello")
		r, err := TryEncryptValue("test", v)

		assert.NoError(t, err)
		assert.NotEqual(t, v, r)

		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Secondary: pr,
		})

		dr, err := TryDecryptValue("test", r)
		assert.NoError(t, err)
		assert.Equal(t, v, dr)
	})

	t.Run("state store with AES256 primary key, base64 string value encrypted and decrypted successfully", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}

		bytes := make([]byte, 32)
		rand.Read(bytes)

		key := hex.EncodeToString(bytes)

		pr := Key{
			Name: "primary",
			Key:  key,
		}

		gcm, _ := createCipher(pr, AES256Algorithm)
		pr.gcm = gcm

		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		AddEncryptedStateStore("test", ComponentEncryptionKeys{
			Primary: pr,
		})

		v := []byte("hello")
		s := base64.StdEncoding.EncodeToString(v)
		r, err := TryEncryptValue("test", []byte(s))

		assert.NoError(t, err)
		assert.NotEqual(t, v, r)

		dr, err := TryDecryptValue("test", r)
		assert.NoError(t, err)
		assert.Equal(t, []byte(s), dr)
	})
}

func TestEncryptedStateStore(t *testing.T) {
	t.Run("store supports encryption", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		AddEncryptedStateStore("test", ComponentEncryptionKeys{})
		ok := EncryptedStateStore("test")

		assert.True(t, ok)
	})

	t.Run("store doesn't support encryption", func(t *testing.T) {
		encryptedStateStores = map[string]ComponentEncryptionKeys{}
		ok := EncryptedStateStore("test")

		assert.False(t, ok)
	})
}
