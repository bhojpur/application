package injector

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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"

	"github.com/bhojpur/application/pkg/client/clientset/versioned/fake"
)

const (
	appPort = "5000"
)

func TestConfigCorrectValues(t *testing.T) {
	i := NewInjector(nil, Config{
		TLSCertFile:            "a",
		TLSKeyFile:             "b",
		SidecarImage:           "c",
		SidecarImagePullPolicy: "d",
		Namespace:              "e",
	}, nil, nil)

	injector := i.(*injector)
	assert.Equal(t, "a", injector.config.TLSCertFile)
	assert.Equal(t, "b", injector.config.TLSKeyFile)
	assert.Equal(t, "c", injector.config.SidecarImage)
	assert.Equal(t, "d", injector.config.SidecarImagePullPolicy)
	assert.Equal(t, "e", injector.config.Namespace)
}

func TestGetConfig(t *testing.T) {
	m := map[string]string{appConfigKey: "config1"}
	c := getConfig(m)
	assert.Equal(t, "config1", c)
}

func TestGetProfiling(t *testing.T) {
	t.Run("missing annotation", func(t *testing.T) {
		m := map[string]string{}
		e := profilingEnabled(m)
		assert.Equal(t, e, false)
	})

	t.Run("enabled", func(t *testing.T) {
		m := map[string]string{appEnableProfilingKey: "yes"}
		e := profilingEnabled(m)
		assert.Equal(t, e, true)
	})

	t.Run("disabled", func(t *testing.T) {
		m := map[string]string{appEnableProfilingKey: "false"}
		e := profilingEnabled(m)
		assert.Equal(t, e, false)
	})
	m := map[string]string{appConfigKey: "config1"}
	c := getConfig(m)
	assert.Equal(t, "config1", c)
}

func TestGetAppPort(t *testing.T) {
	t.Run("valid port", func(t *testing.T) {
		m := map[string]string{appAppPortKey: "3000"}
		p, err := getAppPort(m)
		assert.Nil(t, err)
		assert.Equal(t, int32(3000), p)
	})

	t.Run("invalid port", func(t *testing.T) {
		m := map[string]string{appAppPortKey: "a"}
		p, err := getAppPort(m)
		assert.NotNil(t, err)
		assert.Equal(t, int32(-1), p)
	})
}

func TestGetProtocol(t *testing.T) {
	t.Run("valid grpc protocol", func(t *testing.T) {
		m := map[string]string{appAppProtocolKey: "grpc"}
		p := getProtocol(m)
		assert.Equal(t, "grpc", p)
	})

	t.Run("valid http protocol", func(t *testing.T) {
		m := map[string]string{appAppProtocolKey: "http"}
		p := getProtocol(m)
		assert.Equal(t, "http", p)
	})

	t.Run("get default http protocol", func(t *testing.T) {
		m := map[string]string{}
		p := getProtocol(m)
		assert.Equal(t, "http", p)
	})
}

func TestGetAppID(t *testing.T) {
	t.Run("get app id", func(t *testing.T) {
		m := map[string]string{appIDKey: "app"}
		pod := corev1.Pod{}
		pod.Annotations = m
		id := getAppID(pod)
		assert.Equal(t, "app", id)
	})

	t.Run("get pod id", func(t *testing.T) {
		pod := corev1.Pod{}
		pod.ObjectMeta.Name = "pod"
		id := getAppID(pod)
		assert.Equal(t, "pod", id)
	})
}

func TestLogLevel(t *testing.T) {
	t.Run("empty log level - get default", func(t *testing.T) {
		m := map[string]string{}
		logLevel := getLogLevel(m)
		assert.Equal(t, "info", logLevel)
	})

	t.Run("error log level", func(t *testing.T) {
		m := map[string]string{appLogLevel: "error"}
		logLevel := getLogLevel(m)
		assert.Equal(t, "error", logLevel)
	})
}

func TestMaxConcurrency(t *testing.T) {
	t.Run("empty max concurrency - should be -1", func(t *testing.T) {
		m := map[string]string{}
		maxConcurrency, err := getMaxConcurrency(m)
		assert.Nil(t, err)
		assert.Equal(t, int32(-1), maxConcurrency)
	})

	t.Run("invalid max concurrency - should be -1", func(t *testing.T) {
		m := map[string]string{appAppMaxConcurrencyKey: "invalid"}
		_, err := getMaxConcurrency(m)
		assert.NotNil(t, err)
	})

	t.Run("valid max concurrency - should be 10", func(t *testing.T) {
		m := map[string]string{appAppMaxConcurrencyKey: "10"}
		maxConcurrency, err := getMaxConcurrency(m)
		assert.Nil(t, err)
		assert.Equal(t, int32(10), maxConcurrency)
	})
}

func TestGetServiceAddress(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		clusterDomain string
		port          int
		expect        string
	}{
		{
			port:          80,
			name:          "a",
			namespace:     "b",
			clusterDomain: "cluster.local",
			expect:        "a.b.svc.cluster.local:80",
		},
		{
			port:          50001,
			name:          "app",
			namespace:     "default",
			clusterDomain: "selfdefine.domain",
			expect:        "app.default.svc.selfdefine.domain:50001",
		},
	}
	for _, tc := range testCases {
		dns := getServiceAddress(tc.name, tc.namespace, tc.clusterDomain, tc.port)
		assert.Equal(t, tc.expect, dns)
	}
}

func TestGetMetricsPort(t *testing.T) {
	t.Run("metrics port override", func(t *testing.T) {
		m := map[string]string{appMetricsPortKey: "5050"}
		pod := corev1.Pod{}
		pod.Annotations = m
		p := getMetricsPort(pod.Annotations)
		assert.Equal(t, 5050, p)
	})
	t.Run("invalid metrics port override", func(t *testing.T) {
		m := map[string]string{appMetricsPortKey: "abc"}
		pod := corev1.Pod{}
		pod.Annotations = m
		p := getMetricsPort(pod.Annotations)
		assert.Equal(t, defaultMetricsPort, p)
	})
	t.Run("no metrics port defined", func(t *testing.T) {
		pod := corev1.Pod{}
		p := getMetricsPort(pod.Annotations)
		assert.Equal(t, defaultMetricsPort, p)
	})
}

func TestGetContainer(t *testing.T) {
	annotations := map[string]string{}
	annotations[appConfigKey] = "config"
	annotations[appAppPortKey] = appPort

	c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")

	assert.NotNil(t, c)
	assert.Equal(t, "image", c.Image)
}

func TestSidecarResourceLimits(t *testing.T) {
	t.Run("with limits", func(t *testing.T) {
		annotations := map[string]string{}
		annotations[appConfigKey] = "config1"
		annotations[appAppPortKey] = appPort
		annotations[appLogAsJSON] = "true"
		annotations[appCPULimitKey] = "100m"
		annotations[appMemoryLimitKey] = "1Gi"

		c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")
		assert.NotNil(t, c)
		assert.Equal(t, "100m", c.Resources.Limits.Cpu().String())
		assert.Equal(t, "1Gi", c.Resources.Limits.Memory().String())
	})

	t.Run("with requests", func(t *testing.T) {
		annotations := map[string]string{}
		annotations[appConfigKey] = "config1"
		annotations[appAppPortKey] = appPort
		annotations[appLogAsJSON] = "true"
		annotations[appCPURequestKey] = "100m"
		annotations[appMemoryRequestKey] = "1Gi"

		c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")
		assert.NotNil(t, c)
		assert.Equal(t, "100m", c.Resources.Requests.Cpu().String())
		assert.Equal(t, "1Gi", c.Resources.Requests.Memory().String())
	})

	t.Run("no limits", func(t *testing.T) {
		annotations := map[string]string{}
		annotations[appConfigKey] = "config1"
		annotations[appAppPortKey] = appPort
		annotations[appLogAsJSON] = "true"

		c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")
		assert.NotNil(t, c)
		assert.Len(t, c.Resources.Limits, 0)
	})
}

func TestGetAppIDFromRequest(t *testing.T) {
	t.Run("can handle nil", func(t *testing.T) {
		appID := getAppIDFromRequest(nil)
		assert.Equal(t, "", appID)
	})

	t.Run("can handle empty admissionrequest object", func(t *testing.T) {
		fakeReq := &v1.AdmissionRequest{}
		appID := getAppIDFromRequest(fakeReq)
		assert.Equal(t, "", appID)
	})

	t.Run("can get correct appID", func(t *testing.T) {
		fakePod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"bhojpur.net/app-id": "fakeID",
				},
			},
		}
		rawBytes, _ := json.Marshal(fakePod)
		fakeReq := &v1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: rawBytes,
			},
		}
		appID := getAppIDFromRequest(fakeReq)
		assert.Equal(t, "fakeID", appID)
	})
}

func TestGetResourceRequirements(t *testing.T) {
	t.Run("no resource requirements", func(t *testing.T) {
		r, err := getResourceRequirements(nil)
		assert.Nil(t, err)
		assert.Nil(t, r)
	})

	t.Run("valid resource limits", func(t *testing.T) {
		a := map[string]string{appCPULimitKey: "100m", appMemoryLimitKey: "1Gi"}
		r, err := getResourceRequirements(a)
		assert.Nil(t, err)
		assert.Equal(t, "100m", r.Limits.Cpu().String())
		assert.Equal(t, "1Gi", r.Limits.Memory().String())
	})

	t.Run("invalid cpu limit", func(t *testing.T) {
		a := map[string]string{appCPULimitKey: "cpu", appMemoryLimitKey: "1Gi"}
		r, err := getResourceRequirements(a)
		assert.NotNil(t, err)
		assert.Nil(t, r)
	})

	t.Run("invalid memory limit", func(t *testing.T) {
		a := map[string]string{appCPULimitKey: "100m", appMemoryLimitKey: "memory"}
		r, err := getResourceRequirements(a)
		assert.NotNil(t, err)
		assert.Nil(t, r)
	})

	t.Run("valid resource requests", func(t *testing.T) {
		a := map[string]string{appCPURequestKey: "100m", appMemoryRequestKey: "1Gi"}
		r, err := getResourceRequirements(a)
		assert.Nil(t, err)
		assert.Equal(t, "100m", r.Requests.Cpu().String())
		assert.Equal(t, "1Gi", r.Requests.Memory().String())
	})

	t.Run("invalid cpu request", func(t *testing.T) {
		a := map[string]string{appCPURequestKey: "cpu", appMemoryRequestKey: "1Gi"}
		r, err := getResourceRequirements(a)
		assert.NotNil(t, err)
		assert.Nil(t, r)
	})

	t.Run("invalid memory request", func(t *testing.T) {
		a := map[string]string{appCPURequestKey: "100m", appMemoryRequestKey: "memory"}
		r, err := getResourceRequirements(a)
		assert.NotNil(t, err)
		assert.Nil(t, r)
	})
}

func TestAPITokenSecret(t *testing.T) {
	t.Run("secret exists", func(t *testing.T) {
		annotations := map[string]string{}
		annotations[appAPITokenSecret] = "secret"

		s := getAPITokenSecret(annotations)
		assert.NotNil(t, s)
	})

	t.Run("secret empty", func(t *testing.T) {
		annotations := map[string]string{}
		annotations[appAPITokenSecret] = ""

		s := getAPITokenSecret(annotations)
		assert.Equal(t, "", s)
	})
}

func TestAppSSL(t *testing.T) {
	t.Run("ssl enabled", func(t *testing.T) {
		annotations := map[string]string{
			appAppSSLKey: "true",
		}
		s := appSSLEnabled(annotations)
		assert.True(t, s)
	})

	t.Run("ssl disabled", func(t *testing.T) {
		annotations := map[string]string{
			appAppSSLKey: "false",
		}
		s := appSSLEnabled(annotations)
		assert.False(t, s)
	})

	t.Run("ssl not specified", func(t *testing.T) {
		annotations := map[string]string{}
		s := appSSLEnabled(annotations)
		assert.False(t, s)
	})

	t.Run("get sidecar container enabled", func(t *testing.T) {
		annotations := map[string]string{
			appAppSSLKey: "true",
		}
		c, _ := getSidecarContainer(annotations, "app", "image", "", "ns", "a", "b", nil, "", "", "", "", false, "")
		found := false
		for _, a := range c.Args {
			if a == "--app-ssl" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("get sidecar container disabled", func(t *testing.T) {
		annotations := map[string]string{
			appAppSSLKey: "false",
		}
		c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")
		for _, a := range c.Args {
			if a == "--app-ssl" {
				t.FailNow()
			}
		}
	})

	t.Run("get sidecar container not specified", func(t *testing.T) {
		annotations := map[string]string{}
		c, _ := getSidecarContainer(annotations, "app", "image", "Always", "ns", "a", "b", nil, "", "", "", "", false, "")
		for _, a := range c.Args {
			if a == "--app-ssl" {
				t.FailNow()
			}
		}
	})
}

func TestHandleRequest(t *testing.T) {
	authID := "test-auth-id"

	i := NewInjector([]string{authID}, Config{
		TLSCertFile:  "test-cert",
		TLSKeyFile:   "test-key",
		SidecarImage: "test-image",
		Namespace:    "test-ns",
	}, fake.NewSimpleClientset(), kubernetesfake.NewSimpleClientset())
	injector := i.(*injector)

	podBytes, _ := json.Marshal(corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-app",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				"app": "test-app",
			},
			Annotations: map[string]string{
				"bhojpur.net/enabled":  "true",
				"bhojpur.net/app-id":   "test-app",
				"bhojpur.net/app-port": "3000",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "docker.io/app:latest",
				},
			},
		},
	})

	testCases := []struct {
		testName         string
		request          v1.AdmissionReview
		contentType      string
		expectStatusCode int
		expectPatched    bool
	}{
		{
			"TestSidecarInjectSuccess",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						Groups: []string{systemGroup},
					},
					Object: runtime.RawExtension{Raw: podBytes},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			true,
		},
		{
			"TestSidecarInjectWrongContentType",
			v1.AdmissionReview{},
			runtime.ContentTypeYAML,
			http.StatusUnsupportedMediaType,
			true,
		},
		{
			"TestSidecarInjectInvalidKind",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Deployment"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						Groups: []string{systemGroup},
					},
					Object: runtime.RawExtension{Raw: podBytes},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			false,
		},
		{
			"TestSidecarInjectGroupsNotContains",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						Groups: []string{"system:kubelet"},
					},
					Object: runtime.RawExtension{Raw: podBytes},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			false,
		},
		{
			"TestSidecarInjectUIDContains",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						UID: authID,
					},
					Object: runtime.RawExtension{Raw: podBytes},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			true,
		},
		{
			"TestSidecarInjectUIDNotContains",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						UID: "auth-id-123",
					},
					Object: runtime.RawExtension{Raw: podBytes},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			false,
		},
		{
			"TestSidecarInjectEmptyPod",
			v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					UID:       uuid.NewUUID(),
					Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					Name:      "test-app",
					Namespace: "test-ns",
					Operation: "CREATE",
					UserInfo: authenticationv1.UserInfo{
						Groups: []string{systemGroup},
					},
					Object: runtime.RawExtension{Raw: nil},
				},
			},
			runtime.ContentTypeJSON,
			http.StatusOK,
			false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(injector.handleRequest))
	defer ts.Close()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			requestBytes, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			resp, err := http.Post(ts.URL, tc.contentType, bytes.NewBuffer(requestBytes))
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var ar v1.AdmissionReview
				err = json.Unmarshal(body, &ar)
				assert.NoError(t, err)

				assert.Equal(t, tc.expectPatched, len(ar.Response.Patch) > 0)
			}
		})
	}
}
