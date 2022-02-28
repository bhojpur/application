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
	"bytes"
	b64 "encoding/base64"

	"github.com/pkg/errors"
)

var encryptedStateStores = map[string]ComponentEncryptionKeys{}

const (
	separator = "||"
)

// AddEncryptedStateStore adds an encrypted state store and an associated encryption key to a list.
func AddEncryptedStateStore(storeName string, keys ComponentEncryptionKeys) bool {
	if _, ok := encryptedStateStores[storeName]; ok {
		return false
	}

	encryptedStateStores[storeName] = keys
	return true
}

// EncryptedStateStore returns a bool that indicates if a state stores supports encryption.
func EncryptedStateStore(storeName string) bool {
	_, ok := encryptedStateStores[storeName]
	return ok
}

// TryEncryptValue will try to encrypt a byte array if the state store has associated encryption keys.
// The function will append the name of the key to the value for later extraction.
// If no encryption keys exist, the function will return the bytes unmodified.
func TryEncryptValue(storeName string, value []byte) ([]byte, error) {
	keys := encryptedStateStores[storeName]
	enc, err := encrypt(value, keys.Primary, AES256Algorithm)
	if err != nil {
		return value, err
	}

	sEnc := b64.StdEncoding.EncodeToString(enc) + separator + keys.Primary.Name
	return []byte(sEnc), nil
}

// TryDecryptValue will try to decrypt a byte array if the state store has associated encryption keys.
// If no encryption keys exist, the function will return the bytes unmodified.
func TryDecryptValue(storeName string, value []byte) ([]byte, error) {
	keys := encryptedStateStores[storeName]
	// extract the decryption key that should be appended to the value
	ind := bytes.LastIndex(value, []byte(separator))
	keyName := string(value[ind+len(separator):])

	if len(keyName) == 0 {
		return value, errors.Errorf("could not decrypt data for state store %s: encryption key name not found on record", storeName)
	}

	var key Key

	if keys.Primary.Name == keyName {
		key = keys.Primary
	} else if keys.Secondary.Name == keyName {
		key = keys.Secondary
	}

	return decrypt(value[:ind], key, AES256Algorithm)
}
