package actors

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
	"testing"
	"time"

	app_config "github.com/bhojpur/application/pkg/config"

	"github.com/stretchr/testify/assert"
)

const (
	HostAddress      = "host"
	AppID            = "testApp"
	PlacementAddress = "placement"
	Port             = 5000
	Namespace        = "default"
)

func TestDefaultConfigValuesSet(t *testing.T) {
	appConfig := app_config.ApplicationConfig{Entities: []string{"actor1"}}
	config := NewConfig(HostAddress, AppID, []string{PlacementAddress}, Port, Namespace, appConfig)

	assert.Equal(t, HostAddress, config.HostAddress)
	assert.Equal(t, AppID, config.AppID)
	assert.Contains(t, config.PlacementAddresses, PlacementAddress)
	assert.Equal(t, Port, config.Port)
	assert.Equal(t, Namespace, config.Namespace)
	assert.NotNil(t, config.ActorIdleTimeout)
	assert.NotNil(t, config.ActorDeactivationScanInterval)
	assert.NotNil(t, config.DrainOngoingCallTimeout)
	assert.NotNil(t, config.DrainRebalancedActors)
}

func TestPerActorTypeConfigurationValues(t *testing.T) {
	appConfig := app_config.ApplicationConfig{
		Entities:                   []string{"actor1", "actor2", "actor3", "actor4"},
		ActorIdleTimeout:           "1s",
		ActorScanInterval:          "2s",
		DrainOngoingCallTimeout:    "5s",
		DrainRebalancedActors:      true,
		RemindersStoragePartitions: 1,
		EntityConfigs: []app_config.EntityConfig{
			{
				Entities:                []string{"actor1", "actor2"},
				ActorIdleTimeout:        "60s",
				DrainOngoingCallTimeout: "300s",
				DrainRebalancedActors:   false,
			},
			{
				Entities:                []string{"actor3"},
				ActorIdleTimeout:        "5s",
				DrainOngoingCallTimeout: "1s",
				DrainRebalancedActors:   true,
				Reentrancy: app_config.ReentrancyConfig{
					Enabled: true,
				},
				RemindersStoragePartitions: 10,
			},
		},
	}
	config := NewConfig(HostAddress, AppID, []string{PlacementAddress}, Port, Namespace, appConfig)

	// Check the base level items.
	assert.Equal(t, HostAddress, config.HostAddress)
	assert.Equal(t, AppID, config.AppID)
	assert.Contains(t, config.PlacementAddresses, PlacementAddress)
	assert.Equal(t, Port, config.Port)
	assert.Equal(t, Namespace, config.Namespace)
	assert.Equal(t, time.Second, config.ActorIdleTimeout)
	assert.Equal(t, time.Second*2, config.ActorDeactivationScanInterval)
	assert.Equal(t, time.Second*5, config.DrainOngoingCallTimeout)
	assert.True(t, config.DrainRebalancedActors)

	// Check the specific actors.
	assert.Equal(t, time.Second*60, config.GetIdleTimeoutForType("actor1"))
	assert.Equal(t, time.Second*300, config.GetDrainOngoingTimeoutForType("actor1"))
	assert.False(t, config.GetDrainRebalancedActorsForType("actor1"))
	assert.False(t, config.GetReentrancyForType("actor1").Enabled)
	assert.Equal(t, 0, config.GetRemindersPartitionCountForType("actor1"))
	assert.Equal(t, time.Second*60, config.GetIdleTimeoutForType("actor2"))
	assert.Equal(t, time.Second*300, config.GetDrainOngoingTimeoutForType("actor2"))
	assert.False(t, config.GetDrainRebalancedActorsForType("actor2"))
	assert.False(t, config.GetReentrancyForType("actor2").Enabled)
	assert.Equal(t, 0, config.GetRemindersPartitionCountForType("actor2"))

	assert.Equal(t, time.Second*5, config.GetIdleTimeoutForType("actor3"))
	assert.Equal(t, time.Second, config.GetDrainOngoingTimeoutForType("actor3"))
	assert.True(t, config.GetDrainRebalancedActorsForType("actor3"))
	assert.True(t, config.GetReentrancyForType("actor3").Enabled)
	assert.Equal(t, 10, config.GetRemindersPartitionCountForType("actor3"))

	assert.Equal(t, time.Second, config.GetIdleTimeoutForType("actor4"))
	assert.Equal(t, time.Second*5, config.GetDrainOngoingTimeoutForType("actor4"))
	assert.True(t, config.GetDrainRebalancedActorsForType("actor4"))
	assert.False(t, config.GetReentrancyForType("actor4").Enabled)
	assert.Equal(t, 1, config.GetRemindersPartitionCountForType("actor4"))
}

func TestOnlyHostedActorTypesAreIncluded(t *testing.T) {
	appConfig := app_config.ApplicationConfig{
		Entities:                   []string{"actor1", "actor2"},
		ActorIdleTimeout:           "1s",
		ActorScanInterval:          "2s",
		DrainOngoingCallTimeout:    "5s",
		DrainRebalancedActors:      true,
		RemindersStoragePartitions: 1,
		EntityConfigs: []app_config.EntityConfig{
			{
				Entities:                []string{"actor1", "actor2"},
				ActorIdleTimeout:        "60s",
				DrainOngoingCallTimeout: "300s",
				DrainRebalancedActors:   false,
			},
			{
				Entities:                []string{"actor3"},
				ActorIdleTimeout:        "5s",
				DrainOngoingCallTimeout: "1s",
				DrainRebalancedActors:   true,
				Reentrancy: app_config.ReentrancyConfig{
					Enabled: true,
				},
				RemindersStoragePartitions: 10,
			},
		},
	}

	config := NewConfig(HostAddress, AppID, []string{PlacementAddress}, Port, Namespace, appConfig)

	assert.Contains(t, config.EntityConfigs, "actor1")
	assert.Contains(t, config.EntityConfigs, "actor2")
	assert.NotContains(t, config.EntityConfigs, "actor3")
}
