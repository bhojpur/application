package handlers

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
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	app_testing "github.com/bhojpur/application/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNewAppHandler(t *testing.T) {
	d := getTestAppHandler()
	assert.True(t, d != nil)
}

func TestGetAppID(t *testing.T) {
	testAppHandler := getTestAppHandler()
	t.Run("WithValidId", func(t *testing.T) {
		// Arrange
		expected := "test_id"
		deployment := getDeployment(expected, "true")

		// Act
		got := testAppHandler.getAppID(deployment)

		// Assert
		assert.Equal(t, expected, got)
	})

	t.Run("WithEmptyId", func(t *testing.T) {
		// Arrange
		expected := ""
		deployment := getDeployment(expected, "true")

		// Act
		got := testAppHandler.getAppID(deployment)

		// Assert
		assert.Equal(t, expected, got)
	})
}

func TestIsAnnotatedForApp(t *testing.T) {
	testAppHandler := getTestAppHandler()
	t.Run("Enabled", func(t *testing.T) {
		// Arrange
		expected := "true"
		deployment := getDeployment("test_id", expected)

		// Act
		got := testAppHandler.isAnnotatedForApp(deployment)

		// Assert
		assert.True(t, got)
	})

	t.Run("Disabled", func(t *testing.T) {
		// Arrange
		expected := "false"
		deployment := getDeployment("test_id", expected)

		// Act
		got := testAppHandler.isAnnotatedForApp(deployment)

		// Assert
		assert.False(t, got)
	})

	t.Run("Invalid", func(t *testing.T) {
		// Arrange
		expected := "0"
		deployment := getDeployment("test_id", expected)

		// Act
		got := testAppHandler.isAnnotatedForApp(deployment)

		// Assert
		assert.False(t, got)
	})
}

func TestAppService(t *testing.T) {
	t.Run("invalid empty app id", func(t *testing.T) {
		d := getDeployment("", "true")
		err := getTestAppHandler().ensureAppServicePresent(context.TODO(), "default", d)
		assert.Error(t, err)
	})

	t.Run("invalid char app id", func(t *testing.T) {
		d := getDeployment("myapp@", "true")
		err := getTestAppHandler().ensureAppServicePresent(context.TODO(), "default", d)
		assert.Error(t, err)
	})
}

func TestCreateAppServiceAppIDAndMetricsSettings(t *testing.T) {
	testAppHandler := getTestAppHandler()
	ctx := context.Background()
	myAppService := types.NamespacedName{
		Namespace: "test",
		Name:      "test",
	}
	deployment := getDeployment("test", "true")
	deployment.GetTemplateAnnotations()[appMetricsPortKey] = "12345"

	service := testAppHandler.createAppServiceValues(ctx, myAppService, deployment, "test")
	require.NotNil(t, service)
	assert.Equal(t, "test", service.ObjectMeta.Annotations[appIDAnnotationKey])
	assert.Equal(t, "true", service.ObjectMeta.Annotations["prometheus.io/scrape"])
	assert.Equal(t, "12345", service.ObjectMeta.Annotations["prometheus.io/port"])
	assert.Equal(t, "/", service.ObjectMeta.Annotations["prometheus.io/path"])

	deployment.GetTemplateAnnotations()[appEnableMetricsKey] = "false"

	service = testAppHandler.createAppServiceValues(ctx, myAppService, deployment, "test")
	require.NotNil(t, service)
	assert.Equal(t, "test", service.ObjectMeta.Annotations[appIDAnnotationKey])
	assert.Equal(t, "", service.ObjectMeta.Annotations["prometheus.io/scrape"])
	assert.Equal(t, "", service.ObjectMeta.Annotations["prometheus.io/port"])
	assert.Equal(t, "", service.ObjectMeta.Annotations["prometheus.io/path"])
}

func TestGetMetricsPort(t *testing.T) {
	testAppHandler := getTestAppHandler()
	t.Run("metrics port override", func(t *testing.T) {
		// Arrange
		deployment := getDeploymentWithMetricsPortAnnotation("test_id", "true", "5050")

		// Act
		p := testAppHandler.getMetricsPort(deployment)

		// Assert
		assert.Equal(t, 5050, p)
	})
	t.Run("invalid metrics port override", func(t *testing.T) {
		// Arrange
		deployment := getDeploymentWithMetricsPortAnnotation("test_id", "true", "abc")

		// Act
		p := testAppHandler.getMetricsPort(deployment)

		// Assert
		assert.Equal(t, defaultMetricsPort, p)
	})
	t.Run("no metrics port override", func(t *testing.T) {
		// Arrange
		deployment := getDeployment("test_id", "true")

		// Act
		p := testAppHandler.getMetricsPort(deployment)

		// Assert
		assert.Equal(t, defaultMetricsPort, p)
	})
}

func TestWrapper(t *testing.T) {
	deploymentWrapper := getDeployment("test_id", "true")
	statefulsetWrapper := getStatefulSet("test_id", "true")

	t.Run("get match labal from wrapper", func(t *testing.T) {
		assert.Equal(t, "test", deploymentWrapper.GetMatchLabels()["app"])
		assert.Equal(t, "test", statefulsetWrapper.GetMatchLabels()["app"])
	})

	t.Run("get annotations from wrapper", func(t *testing.T) {
		assert.Equal(t, "test_id", deploymentWrapper.GetTemplateAnnotations()[appIDAnnotationKey])
		assert.Equal(t, "test_id", statefulsetWrapper.GetTemplateAnnotations()[appIDAnnotationKey])
	})

	t.Run("get object from wrapper", func(t *testing.T) {
		assert.Equal(t, reflect.TypeOf(deploymentWrapper.GetObject()), reflect.TypeOf(&appsv1.Deployment{}))
		assert.Equal(t, reflect.TypeOf(statefulsetWrapper.GetObject()), reflect.TypeOf(&appsv1.StatefulSet{}))
		assert.NotEqual(t, reflect.TypeOf(statefulsetWrapper.GetObject()), reflect.TypeOf(&appsv1.Deployment{}))
		assert.NotEqual(t, reflect.TypeOf(deploymentWrapper.GetObject()), reflect.TypeOf(&appsv1.StatefulSet{}))
	})
}

func TestInit(t *testing.T) {
	mgr := app_testing.NewMockManager()

	_ = scheme.AddToScheme(mgr.GetScheme())

	handler := NewAppHandler(mgr)

	t.Run("test init Bhojpur Application handler", func(t *testing.T) {
		assert.NotNil(t, handler)

		err := handler.Init()

		assert.Nil(t, err)

		assert.Equal(t, 2, len(mgr.GetRunnables()))

		srv := &corev1.Service{}
		val := mgr.GetIndexerFunc(&corev1.Service{})(srv)
		assert.Nil(t, val)

		trueA := true
		srv = &corev1.Service{
			ObjectMeta: meta_v1.ObjectMeta{
				OwnerReferences: []meta_v1.OwnerReference{
					{
						Name:       "TestName",
						Controller: &trueA,
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
			},
		}

		val = mgr.GetIndexerFunc(&corev1.Service{})(srv)
		assert.Equal(t, []string{"TestName"}, val)
	})

	t.Run("test wrapper", func(t *testing.T) {
		deploymentCtl := mgr.GetRunnables()[0]
		statefulsetCtl := mgr.GetRunnables()[1]

		// the runnable is sigs.k8s.io/controller-runtime/pkg/internal/controller.Controller
		reconciler := reflect.Indirect(reflect.ValueOf(deploymentCtl)).FieldByName("Do").Interface().(*Reconciler)

		wrapper := reconciler.newWrapper()

		assert.NotNil(t, wrapper)

		assert.Equal(t, reflect.TypeOf(&appsv1.Deployment{}), reflect.TypeOf(wrapper.GetObject()))

		reconciler = reflect.Indirect(reflect.ValueOf(statefulsetCtl)).FieldByName("Do").Interface().(*Reconciler)

		wrapper = reconciler.newWrapper()

		assert.NotNil(t, wrapper)

		assert.Equal(t, reflect.TypeOf(&appsv1.StatefulSet{}), reflect.TypeOf(wrapper.GetObject()))
	})
}

func getDeploymentWithMetricsPortAnnotation(appID string, appEnabled string, metricsPort string) ObjectWrapper {
	d := getDeployment(appID, appEnabled)
	d.GetTemplateAnnotations()[appMetricsPortKey] = metricsPort
	return d
}

func getDeployment(appID string, appEnabled string) ObjectWrapper {
	// Arrange
	metadata := meta_v1.ObjectMeta{
		Name:   "app",
		Labels: map[string]string{"app": "test_app"},
		Annotations: map[string]string{
			appIDAnnotationKey:      appID,
			appEnabledAnnotationKey: appEnabled,
			appEnableMetricsKey:     "true",
		},
	}

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metadata,
	}

	deployment := appsv1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "app",
		},

		Spec: appsv1.DeploymentSpec{
			Template: podTemplateSpec,
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}

	return &DeploymentWrapper{deployment}
}

func getStatefulSet(appID string, appEnabled string) ObjectWrapper {
	// Arrange
	metadata := meta_v1.ObjectMeta{
		Name:   "app",
		Labels: map[string]string{"app": "test_app"},
		Annotations: map[string]string{
			appIDAnnotationKey:      appID,
			appEnabledAnnotationKey: appEnabled,
			appEnableMetricsKey:     "true",
		},
	}

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metadata,
	}

	stratefulset := appsv1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "app",
		},

		Spec: appsv1.StatefulSetSpec{
			Template: podTemplateSpec,
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}

	return &StatefulSetWrapper{
		stratefulset,
	}
}

func getTestAppHandler() *AppHandler {
	return &AppHandler{}
}
