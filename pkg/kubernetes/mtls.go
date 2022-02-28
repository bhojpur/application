package kubernetes

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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bhojpur/application/pkg/utils"
	"github.com/bhojpur/application/pkg/kubernetes/configuration/v1alpha1"
)

const (
	systemConfigName         = "appsystem"
	trustBundleSecretName    = "app-trust-bundle" // nolint:gosec
	warningDaysForCertExpiry = 30                  // in days
)

func IsMTLSEnabled() (bool, error) {
	c, err := getSystemConfig()
	if err != nil {
		return false, err
	}
	return c.Spec.MTLSSpec.Enabled, nil
}

func getSystemConfig() (*v1alpha1.Configuration, error) {
	client, err := AppClient()
	if err != nil {
		return nil, err
	}

	configs, err := client.ConfigurationV1alpha1().Configurations(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
	// This means that the Bhojpur Application Configurations CRD is not installed and
	// therefore no configuration items exist.
	if apierrors.IsNotFound(err) {
		configs = &v1alpha1.ConfigurationList{
			Items: []v1alpha1.Configuration{},
		}
	} else if err != nil {
		return nil, err
	}

	for _, c := range configs.Items {
		if c.GetName() == systemConfigName {
			return &c, nil
		}
	}

	return nil, errors.New("system configuration not found")
}

// ExportTrustChain takes the root cert, issuer cert and issuer key from a k8s cluster and saves them in a given directory.
func ExportTrustChain(outputDir string) error {
	_, err := os.Stat(outputDir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(outputDir, 0755)
		if errDir != nil {
			return err
		}
	}

	secret, err := getTrustChainSecret()
	if err != nil {
		return err
	}

	ca := secret.Data["ca.crt"]
	issuerCert := secret.Data["issuer.crt"]
	issuerKey := secret.Data["issuer.key"]

	err = ioutil.WriteFile(filepath.Join(outputDir, "ca.crt"), ca, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(outputDir, "issuer.crt"), issuerCert, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(outputDir, "issuer.key"), issuerKey, 0600)
	if err != nil {
		return err
	}
	return nil
}

// Check and warn if cert expiry is less than `warningDaysForCertExpiry` days.
func CheckForCertExpiry() {
	expiry, err := Expiry()
	// The intent is to warn for certificate expiry, only when it can be fetched.
	// Do not show any kind of errors with normal command flow.
	if err != nil {
		return
	}
	daysRemaining := int(expiry.Sub(time.Now().UTC()).Hours() / 24)

	if daysRemaining < warningDaysForCertExpiry {
		warningMessage := ""
		switch {
		case daysRemaining == 0:
			warningMessage = "Bhojpur Application runtime root certificate of your Kubernetes cluster expires today."
		case daysRemaining < 0:
			warningMessage = "Bhojpur Application runtime root certificate of your Kubernetes cluster has expired."
		default:
			warningMessage = fmt.Sprintf("Bhojpur Application runtime root certificate of your Kubernetes cluster expires in %v days.", daysRemaining)
		}
		helpMessage := "Please see docs.bhojpur.net for certificate renewal instructions to avoid service interruptions."
		utils.WarningStatusEvent(os.Stdout,
			fmt.Sprintf("%s Expiry date: %s. \n %s", warningMessage, expiry.Format(time.RFC1123), helpMessage))
	}
}

func getTrustChainSecret() (*corev1.Secret, error) {
	_, client, err := GetKubeConfigClient()
	if err != nil {
		return nil, err
	}

	c, err := getSystemConfig()
	if err != nil {
		return nil, err
	}
	res, err := client.CoreV1().Secrets(c.GetNamespace()).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, i := range res.Items {
		if i.GetName() == trustBundleSecretName {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("could not find trust chain secret named %s in namespace %s", trustBundleSecretName, c.GetNamespace())
}

// Expiry returns the expiry time for the root cert.
func Expiry() (*time.Time, error) {
	secret, err := getTrustChainSecret()
	if err != nil {
		return nil, err
	}

	caCrt := secret.Data["ca.crt"]
	block, _ := pem.Decode(caCrt)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &cert.NotAfter, nil
}