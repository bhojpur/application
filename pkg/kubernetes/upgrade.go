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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	helm "helm.sh/helm/v3/pkg/action"
	"k8s.io/helm/pkg/strvals"

	"github.com/hashicorp/go-version"

	"github.com/bhojpur/application/pkg/utils"
)

const operatorName = "app-operator"

var crds = []string{
	"components",
	"configuration",
	"subscription",
}

var crdsFullResources = []string{
	"components.bhojpur.net",
	"configurations.bhojpur.net",
	"subscriptions.bhojpur.net",
}

type UpgradeConfig struct {
	RuntimeVersion string
	Args           []string
	Timeout        uint
}

func Upgrade(conf UpgradeConfig) error {
	sc, err := NewStatusClient()
	if err != nil {
		return err
	}

	status, err := sc.Status()
	if err != nil {
		return err
	}

	if len(status) == 0 {
		return errors.New("Bhojpur Application is not installed in your cluster")
	}

	var appVersion string
	for _, s := range status {
		if s.Name == operatorName {
			appVersion = s.Version
		}
	}
	utils.InfoStatusEvent(os.Stdout, "Bhojpur Application control plane version %s detected in namespace %s", appVersion, status[0].Namespace)

	helmConf, err := helmConfig(status[0].Namespace)
	if err != nil {
		return err
	}

	appChart, err := appChart(conf.RuntimeVersion, helmConf)
	if err != nil {
		return err
	}

	upgradeClient := helm.NewUpgrade(helmConf)
	upgradeClient.ResetValues = true
	upgradeClient.Namespace = status[0].Namespace
	upgradeClient.CleanupOnFail = true
	upgradeClient.Wait = true
	upgradeClient.Timeout = time.Duration(conf.Timeout) * time.Second

	utils.InfoStatusEvent(os.Stdout, "Starting upgrade...")

	mtls, err := IsMTLSEnabled()
	if err != nil {
		return err
	}

	var vals map[string]interface{}
	var ca []byte
	var issuerCert []byte
	var issuerKey []byte

	if mtls {
		secret, sErr := getTrustChainSecret()
		if sErr != nil {
			return sErr
		}

		ca = secret.Data["ca.crt"]
		issuerCert = secret.Data["issuer.crt"]
		issuerKey = secret.Data["issuer.key"]
	}

	ha := highAvailabilityEnabled(status)
	vals, err = upgradeChartValues(string(ca), string(issuerCert), string(issuerKey), ha, mtls, conf.Args)
	if err != nil {
		return err
	}

	if !isDowngrade(conf.RuntimeVersion, appVersion) {
		err = applyCRDs(fmt.Sprintf("v%s", conf.RuntimeVersion))
		if err != nil {
			return err
		}
	} else {
		utils.InfoStatusEvent(os.Stdout, "Downgrade detected, skipping CRDs.")
	}

	listClient := helm.NewList(helmConf)
	releases, err := listClient.Run()
	if err != nil {
		return err
	}

	var chart string
	for _, r := range releases {
		if r.Chart != nil && strings.Contains(r.Chart.Name(), "bhojpur") {
			chart = r.Name
			break
		}
	}

	if _, err = upgradeClient.Run(chart, appChart, vals); err != nil {
		return err
	}
	return nil
}

func highAvailabilityEnabled(status []StatusOutput) bool {
	for _, s := range status {
		if s.Replicas > 1 {
			return true
		}
	}
	return false
}

func applyCRDs(version string) error {
	for _, crd := range crds {
		url := fmt.Sprintf("https://raw.githubusercontent.com/bhojpur/application/%s/charts/app/crds/%s.yaml", version, crd)
		_, err := utils.RunCmdAndWait("kubectl", "apply", "-f", url)
		if err != nil {
			return err
		}
	}
	return nil
}

func upgradeChartValues(ca, issuerCert, issuerKey string, haMode, mtls bool, args []string) (map[string]interface{}, error) {
	chartVals := map[string]interface{}{}
	globalVals := args

	if mtls && ca != "" && issuerCert != "" && issuerKey != "" {
		globalVals = append(globalVals, fmt.Sprintf("app_sentry.tls.root.certPEM=%s", ca),
			fmt.Sprintf("app_sentry.tls.issuer.certPEM=%s", issuerCert),
			fmt.Sprintf("app_sentry.tls.issuer.keyPEM=%s", issuerKey),
		)
	} else {
		globalVals = append(globalVals, "global.mtls.enabled=false")
	}

	if haMode {
		globalVals = append(globalVals, "global.ha.enabled=true")
	}

	for _, v := range globalVals {
		if err := strvals.ParseInto(v, chartVals); err != nil {
			return nil, err
		}
	}
	return chartVals, nil
}

func isDowngrade(targetVersion, existingVersion string) bool {
	target, _ := version.NewVersion(targetVersion)
	existing, _ := version.NewVersion(existingVersion)

	return target.LessThan(existing)
}
