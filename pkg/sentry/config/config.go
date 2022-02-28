package config

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
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bhojpur/service/pkg/utils/logger"

	scheme "github.com/bhojpur/application/pkg/client/clientset/versioned"
	app_config "github.com/bhojpur/application/pkg/config"
	"github.com/bhojpur/application/pkg/utils"
)

const (
	kubernetesServiceHostEnvVar = "KUBERNETES_SERVICE_HOST"
	kubernetesConfig            = "kubernetes"
	selfHostedConfig            = "selfhosted"
	defaultPort                 = 50001
	defaultWorkloadCertTTL      = time.Hour * 24
	defaultAllowedClockSkew     = time.Minute * 15

	// defaultAppSystemConfigName is the default resource object name for Bhojpur Application System Config.
	defaultAppSystemConfigName = "appsystem"
)

var log = logger.NewLogger("app.sentry.config")

// SentryConfig holds the configuration for the Certificate Authority.
type SentryConfig struct {
	Port             int
	TrustDomain      string
	CAStore          string
	WorkloadCertTTL  time.Duration
	AllowedClockSkew time.Duration
	RootCertPath     string
	IssuerCertPath   string
	IssuerKeyPath    string
}

var configGetters = map[string]func(string) (SentryConfig, error){
	selfHostedConfig: getSelfhostedConfig,
	kubernetesConfig: getKubernetesConfig,
}

// FromConfigName returns a Sentry configuration based on a configuration spec.
// A default configuration is loaded in case of an error.
func FromConfigName(configName string) (SentryConfig, error) {
	var confGetterFn func(string) (SentryConfig, error)

	if IsKubernetesHosted() {
		confGetterFn = configGetters[kubernetesConfig]
	} else {
		confGetterFn = configGetters[selfHostedConfig]
	}

	conf, err := confGetterFn(configName)
	if err != nil {
		err = errors.Wrapf(err, "loading default config. couldn't find config name: %s", configName)
		conf = getDefaultConfig()
	}

	printConfig(conf)
	return conf, err
}

func printConfig(config SentryConfig) {
	caStore := "default"
	if config.CAStore != "" {
		caStore = config.CAStore
	}

	log.Infof("configuration: [port]: %v, [ca store]: %s, [allowed clock skew]: %s, [workload cert ttl]: %s",
		config.Port, caStore, config.AllowedClockSkew.String(), config.WorkloadCertTTL.String())
}

func IsKubernetesHosted() bool {
	return os.Getenv(kubernetesServiceHostEnvVar) != ""
}

func getDefaultConfig() SentryConfig {
	return SentryConfig{
		Port:             defaultPort,
		WorkloadCertTTL:  defaultWorkloadCertTTL,
		AllowedClockSkew: defaultAllowedClockSkew,
	}
}

func getKubernetesConfig(configName string) (SentryConfig, error) {
	defaultConfig := getDefaultConfig()

	kubeConf := utils.GetConfig()
	appClient, err := scheme.NewForConfig(kubeConf)
	if err != nil {
		return defaultConfig, err
	}

	list, err := appClient.ConfigurationV1alpha1().Configurations(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err != nil {
		return defaultConfig, err
	}

	if configName == "" {
		configName = defaultAppSystemConfigName
	}

	for _, i := range list.Items {
		if i.GetName() == configName {
			spec, _ := json.Marshal(i.Spec)

			var configSpec app_config.ConfigurationSpec
			json.Unmarshal(spec, &configSpec)

			conf := app_config.Configuration{
				Spec: configSpec,
			}
			return parseConfiguration(defaultConfig, &conf)
		}
	}
	return defaultConfig, errors.New("config CRD not found")
}

func getSelfhostedConfig(configName string) (SentryConfig, error) {
	defaultConfig := getDefaultConfig()
	appConfig, _, err := app_config.LoadStandaloneConfiguration(configName)
	if err != nil {
		return defaultConfig, err
	}

	if appConfig != nil {
		return parseConfiguration(defaultConfig, appConfig)
	}
	return defaultConfig, nil
}

func parseConfiguration(conf SentryConfig, appConfig *app_config.Configuration) (SentryConfig, error) {
	if appConfig.Spec.MTLSSpec.WorkloadCertTTL != "" {
		d, err := time.ParseDuration(appConfig.Spec.MTLSSpec.WorkloadCertTTL)
		if err != nil {
			return conf, errors.Wrap(err, "error parsing WorkloadCertTTL duration")
		}

		conf.WorkloadCertTTL = d
	}

	if appConfig.Spec.MTLSSpec.AllowedClockSkew != "" {
		d, err := time.ParseDuration(appConfig.Spec.MTLSSpec.AllowedClockSkew)
		if err != nil {
			return conf, errors.Wrap(err, "error parsing AllowedClockSkew duration")
		}

		conf.AllowedClockSkew = d
	}

	return conf, nil
}
