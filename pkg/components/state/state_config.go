package state

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
	"strings"

	"github.com/pkg/errors"
)

const (
	strategyKey = "keyPrefix"

	strategyAppid     = "appid"
	strategyStoreName = "name"
	strategyNone      = "none"
	strategyDefault   = strategyAppid

	appSeparator = "||"
)

var statesConfiguration = map[string]*StoreConfiguration{}

type StoreConfiguration struct {
	keyPrefixStrategy string
}

func SaveStateConfiguration(storeName string, metadata map[string]string) error {
	strategy := metadata[strategyKey]
	strategy = strings.ToLower(strategy)
	if strategy == "" {
		strategy = strategyDefault
	} else {
		err := checkKeyIllegal(metadata[strategyKey])
		if err != nil {
			return err
		}
	}

	statesConfiguration[storeName] = &StoreConfiguration{keyPrefixStrategy: strategy}
	return nil
}

func GetModifiedStateKey(key, storeName, appID string) (string, error) {
	if err := checkKeyIllegal(key); err != nil {
		return "", err
	}
	stateConfiguration := getStateConfiguration(storeName)
	switch stateConfiguration.keyPrefixStrategy {
	case strategyNone:
		return key, nil
	case strategyStoreName:
		return fmt.Sprintf("%s%s%s", storeName, appSeparator, key), nil
	case strategyAppid:
		if appID == "" {
			return key, nil
		}
		return fmt.Sprintf("%s%s%s", appID, appSeparator, key), nil
	default:
		return fmt.Sprintf("%s%s%s", stateConfiguration.keyPrefixStrategy, appSeparator, key), nil
	}
}

func GetOriginalStateKey(modifiedStateKey string) string {
	splits := strings.Split(modifiedStateKey, appSeparator)
	if len(splits) <= 1 {
		return modifiedStateKey
	}
	return splits[1]
}

func getStateConfiguration(storeName string) *StoreConfiguration {
	c := statesConfiguration[storeName]
	if c == nil {
		c = &StoreConfiguration{keyPrefixStrategy: strategyDefault}
		statesConfiguration[storeName] = c
	}

	return c
}

func checkKeyIllegal(key string) error {
	if strings.Contains(key, appSeparator) {
		return errors.Errorf("input key/keyPrefix '%s' can't contain '%s'", key, appSeparator)
	}
	return nil
}
