package operator

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
	"context"
	"encoding/json"
	"os"
	"strings"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Register the k8s client auth
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	subscriptionsapi_v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/subscriptions/v1alpha1"
	subscriptionsapi_v2alpha1 "github.com/bhojpur/application/pkg/kubernetes/subscriptions/v2alpha1"
)

const webhookCAName = "app-webhook-ca"

func RunWebhooks(enableLeaderElection bool) {
	conf, err := ctrl.GetConfig()
	if err != nil {
		log.Fatalf("unable to get controller runtime configuration, err: %s", err)
	}

	mgr, err := ctrl.NewManager(conf, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		Port:                   19443,
		HealthProbeBindAddress: "0",
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "webhooks.bhojpur.net",
	})
	if err != nil {
		log.Fatal(err, "unable to start webhooks")
	}

	/*
		Make sure to set `ENABLE_WEBHOOKS=false` when we run locally.
	*/
	if !strings.EqualFold(os.Getenv("ENABLE_WEBHOOKS"), "false") {
		if err = ctrl.NewWebhookManagedBy(mgr).
			For(&subscriptionsapi_v1alpha1.Subscription{}).
			Complete(); err != nil {
			log.Fatalf("unable to create webhook Subscriptions v1alpha1: %v", err)
		}
		if err = ctrl.NewWebhookManagedBy(mgr).
			For(&subscriptionsapi_v2alpha1.Subscription{}).
			Complete(); err != nil {
			log.Fatalf("unable to create webhook Subscriptions v2alpha1: %v", err)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up ready check: %v", err)
	}

	ctx := ctrl.SetupSignalHandler()

	go patchCRDs(ctx, conf, "subscriptions.bhojpur.net")

	log.Info("starting webhooks")
	if err := mgr.Start(ctx); err != nil {
		log.Fatalf("problem running webhooks: %v", err)
	}
}

func patchCRDs(ctx context.Context, conf *rest.Config, crdNames ...string) {
	client, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Errorf("Could not get Kubernetes API client: %v", err)

		return
	}
	clientSet, err := apiextensionsclient.NewForConfig(conf)
	if err != nil {
		log.Errorf("Could not get API extension client: %v", err)

		return
	}

	crdClient := clientSet.ApiextensionsV1().CustomResourceDefinitions()
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Error("Could not get Bhojpur Application namespace")

		return
	}

	si, err := client.CoreV1().Secrets(namespace).Get(ctx, webhookCAName, v1.GetOptions{})
	if err != nil {
		log.Debugf("Could not get webhook CA: %v", err)
		log.Info("The webhook CA secret was not found. Assuming conversion webhook caBundles are managed manually.")

		return
	}

	caBundle, ok := si.Data["caBundle"]
	if !ok {
		log.Error("Webhook CA secret did not contain 'caBundle'")

		return
	}

	for _, crdName := range crdNames {
		crd, err := crdClient.Get(ctx, crdName, v1.GetOptions{})
		if err != nil {
			log.Errorf("Could not get CRD %q: %v", crdName, err)

			continue
		}

		if crd == nil ||
			crd.Spec.Conversion == nil ||
			crd.Spec.Conversion.Webhook == nil ||
			crd.Spec.Conversion.Webhook.ClientConfig == nil {
			log.Errorf("CRD %q does not have an existing webhook client config. Applying resources of this type will fail.", crdName)

			continue
		}

		if crd.Spec.Conversion.Webhook.ClientConfig.Service != nil &&
			crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace == namespace &&
			crd.Spec.Conversion.Webhook.ClientConfig.CABundle != nil &&
			bytes.Equal(crd.Spec.Conversion.Webhook.ClientConfig.CABundle, caBundle) {
			log.Infof("Conversion webhook for %q is up to date", crdName)

			continue
		}

		// This code mimics:
		// kubectl patch crd "subscriptions.bhojpur.net" --type='json' -p [{'op': 'replace', 'path': '/spec/conversion/webhook/clientConfig/service/namespace', 'value':'${namespace}'},{'op': 'add', 'path': '/spec/conversion/webhook/clientConfig/caBundle', 'value':'${caBundle}'}]"
		type patchValue struct {
			Op    string      `json:"op"`
			Path  string      `json:"path"`
			Value interface{} `json:"value"`
		}
		payload := []patchValue{{
			Op:    "replace",
			Path:  "/spec/conversion/webhook/clientConfig/service/namespace",
			Value: namespace,
		}, {
			Op:    "replace",
			Path:  "/spec/conversion/webhook/clientConfig/caBundle",
			Value: caBundle,
		}}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("Could not marshal webhook spec: %v", err)

			continue
		}
		_, err = crdClient.Patch(ctx, crdName, types.JSONPatchType, payloadJSON, v1.PatchOptions{})
		if err != nil {
			log.Errorf("Failed to patch webhook in CRD %q: %v", crdName, err)

			continue
		}

		log.Infof("Successfully patched webhook in CRD %q", crdName)
	}
}
