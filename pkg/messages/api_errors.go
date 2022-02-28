package messages

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

const (
	// Http.
	ErrNotFound             = "method %q is not found"
	ErrMalformedRequest     = "failed deserializing HTTP body: %s"
	ErrMalformedRequestData = "can't serialize request data field: %s"

	// State.
	ErrStateStoresNotConfigured = "state store is not configured"
	ErrStateStoreNotFound       = "state store %s is not found"
	ErrStateGet                 = "fail to get %s from state store %s: %s"
	ErrStateDelete              = "failed deleting state with key %s: %s"
	ErrStateSave                = "failed saving state in state store %s: %s"
	ErrStateQuery               = "failed query in state store %s: %s"

	// StateTransaction.
	ErrStateStoreNotSupported     = "state store %s doesn't support transaction"
	ErrNotSupportedStateOperation = "operation type %s not supported"
	ErrStateTransaction           = "error while executing state transaction: %s"

	// Binding.
	ErrInvokeOutputBinding = "error when invoke output binding %s: %s"

	// PubSub.
	ErrPubsubNotConfigured      = "no pubsub is configured"
	ErrPubsubEmpty              = "pubsub name is empty"
	ErrPubsubNotFound           = "pubsub %s not found"
	ErrTopicEmpty               = "topic is empty in pubsub %s"
	ErrPubsubCloudEventsSer     = "error when marshalling cloud event envelope for topic %s pubsub %s: %s"
	ErrPubsubPublishMessage     = "error when publish to topic %s in pubsub %s: %s"
	ErrPubsubForbidden          = "topic %s is not allowed for app id %s"
	ErrPubsubCloudEventCreation = "cannot create cloudevent: %s"

	// AppChannel.
	ErrChannelNotFound       = "app channel is not initialized"
	ErrInternalInvokeRequest = "parsing InternalInvokeRequest error: %s"
	ErrChannelInvoke         = "error invoking app channel: %s"

	// Actor.
	ErrActorRuntimeNotFound      = "actor runtime is not configured"
	ErrActorInstanceMissing      = "actor instance is missing"
	ErrActorInvoke               = "error invoke actor method: %s"
	ErrActorReminderCreate       = "error creating actor reminder: %s"
	ErrActorReminderRename       = "error rename actor reminder: %s"
	ErrActorReminderGet          = "error getting actor reminder: %s"
	ErrActorReminderDelete       = "error deleting actor reminder: %s"
	ErrActorTimerCreate          = "error creating actor timer: %s"
	ErrActorTimerDelete          = "error deleting actor timer: %s"
	ErrActorStateGet             = "error getting actor state: %s"
	ErrActorStateTransactionSave = "error saving actor transaction state: %s"

	// Secret.
	ErrSecretStoreNotConfigured = "secret store is not configured"
	ErrSecretStoreNotFound      = "failed finding secret store with key %s"
	ErrPermissionDenied         = "access denied by policy to get %q from %q"
	ErrSecretGet                = "failed getting secret with key %s from secret store %s: %s"
	ErrBulkSecretGet            = "failed getting secrets from secret store %s: %s"

	// DirectMessaging.
	ErrDirectInvoke         = "fail to invoke, id: %s, err: %s"
	ErrDirectInvokeNoAppID  = "failed getting app id either from the URL path or the header app-app-id"
	ErrDirectInvokeMethod   = "invalid method name"
	ErrDirectInvokeNotReady = "invoke API is not ready"

	// Metadata.
	ErrMetadataGet = "failed deserializing metadata: %s"

	// Healthz.
	ErrHealthNotReady = "Bhojpur Application runtime is not ready"

	// Configuration.
	ErrConfigurationStoresNotConfigured = "error configuration stores not configured"
	ErrConfigurationStoreNotFound       = "error configuration stores %s not found"
	ErrConfigurationGet                 = "fail to get %s from Configuration store %s: %s"
	ErrConfigurationSubscribe           = "fail to subscribe %s from Configuration store %s: %s"
)
