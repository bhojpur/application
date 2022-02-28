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
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/operator/monitoring"
	"github.com/bhojpur/application/pkg/validations"
)

const (
	appEnabledAnnotationKey        = "bhojpur.net/enabled"
	appIDAnnotationKey             = "bhojpur.net/app-id"
	appEnableMetricsKey            = "bhojpur.net/enable-metrics"
	appMetricsPortKey              = "bhojpur.net/metrics-port"
	appSidecarHTTPPortName         = "app-http"
	appSidecarAPIGRPCPortName      = "app-grpc"
	appSidecarInternalGRPCPortName = "app-internal"
	appSidecarMetricsPortName      = "app-metrics"
	appSidecarHTTPPort             = 3500
	appSidecarAPIGRPCPort          = 50001
	appSidecarInternalGRPCPort     = 50002
	defaultMetricsEnabled          = true
	defaultMetricsPort             = 9090
	clusterIPNone                  = "None"
	appServiceOwnerField           = ".metadata.controller"
)

var log = logger.NewLogger("app.operator.handlers")

// AppHandler handles the lifetime for Bhojpur Application runtime CRDs.
type AppHandler struct {
	mgr ctrl.Manager

	client.Client
	Scheme *runtime.Scheme
}

type Reconciler struct {
	*AppHandler
	newWrapper func() ObjectWrapper
}

// NewAppHandler returns a new Bhojpur Application runtime handler.
func NewAppHandler(mgr ctrl.Manager) *AppHandler {
	return &AppHandler{
		mgr: mgr,

		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

// Init allows for various startup tasks.
func (h *AppHandler) Init() error {
	if err := h.mgr.GetFieldIndexer().IndexField(
		context.TODO(),
		&corev1.Service{},
		appServiceOwnerField,
		func(rawObj client.Object) []string {
			svc := rawObj.(*corev1.Service)
			owner := meta_v1.GetControllerOf(svc)
			if owner == nil || owner.APIVersion != appsv1.SchemeGroupVersion.String() || (owner.Kind != "Deployment" && owner.Kind != "StatefulSet") {
				return nil
			}
			return []string{owner.Name}
		}); err != nil {
		return err
	}

	if err := ctrl.NewControllerManagedBy(h.mgr).
		For(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 100,
		}).
		Complete(&Reconciler{
			AppHandler: h,
			newWrapper: func() ObjectWrapper {
				return &DeploymentWrapper{}
			},
		}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(h.mgr).
		For(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 100,
		}).
		Complete(&Reconciler{
			AppHandler: h,
			newWrapper: func() ObjectWrapper {
				return &StatefulSetWrapper{}
			},
		})
}

func (h *AppHandler) appServiceName(appID string) string {
	return fmt.Sprintf("%s-app", appID)
}

// Reconcile the expected services for deployments | statefulset annotated for Bhojpur Application.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// var wrapper appsv1.Deployment | appsv1.StatefulSet
	wrapper := r.newWrapper()

	expectedService := false
	if err := r.Get(ctx, req.NamespacedName, wrapper.GetObject()); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debugf("deployment has be deleted, %s", req.NamespacedName)
		} else {
			log.Errorf("unable to get deployment, %s, err: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
	} else {
		if wrapper.GetObject().GetDeletionTimestamp() != nil {
			log.Debugf("deployment is being deleted, %s", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		expectedService = r.isAnnotatedForApp(wrapper)
	}

	if expectedService {
		if err := r.ensureAppServicePresent(ctx, req.Namespace, wrapper); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

func (h *AppHandler) ensureAppServicePresent(ctx context.Context, namespace string, wrapper ObjectWrapper) error {
	appID := h.getAppID(wrapper)
	err := validations.ValidateKubernetesAppID(appID)
	if err != nil {
		return err
	}

	mayAppService := types.NamespacedName{
		Namespace: namespace,
		Name:      h.appServiceName(appID),
	}
	var appSvc corev1.Service
	if err := h.Get(ctx, mayAppService, &appSvc); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debugf("no service for wrapper found, wrapper: %s/%s", namespace, mayAppService.Name)
			return h.createAppService(ctx, mayAppService, wrapper)
		}
		log.Errorf("unable to get service, %s, err: %s", mayAppService, err)
		return err
	}
	return nil
}

func (h *AppHandler) createAppService(ctx context.Context, expectedService types.NamespacedName, wrapper ObjectWrapper) error {
	appID := h.getAppID(wrapper)
	service := h.createAppServiceValues(ctx, expectedService, wrapper, appID)

	if err := ctrl.SetControllerReference(wrapper.GetObject(), service, h.Scheme); err != nil {
		return err
	}
	if err := h.Create(ctx, service); err != nil {
		log.Errorf("unable to create Bhojpur Application service for wrapper, service: %s, err: %s", expectedService, err)
		return err
	}
	log.Debugf("created service: %s", expectedService)
	monitoring.RecordServiceCreatedCount(appID)
	return nil
}

func (h *AppHandler) createAppServiceValues(ctx context.Context, expectedService types.NamespacedName, wrapper ObjectWrapper, appID string) *corev1.Service {
	enableMetrics := h.getEnableMetrics(wrapper)
	metricsPort := h.getMetricsPort(wrapper)
	log.Debugf("enableMetrics: %v", enableMetrics)

	annotations := map[string]string{
		appIDAnnotationKey: appID,
	}

	if enableMetrics {
		annotations["prometheus.io/probe"] = "true"
		annotations["prometheus.io/scrape"] = "true" // WARN: deprecated as of v1.7 please use prometheus.io/probe instead.
		annotations["prometheus.io/port"] = strconv.Itoa(metricsPort)
		annotations["prometheus.io/path"] = "/"
	}

	return &corev1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        expectedService.Name,
			Namespace:   expectedService.Namespace,
			Labels:      map[string]string{appEnabledAnnotationKey: "true"},
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector:  wrapper.GetMatchLabels(),
			ClusterIP: clusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(appSidecarHTTPPort),
					Name:       appSidecarHTTPPortName,
				},
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(appSidecarAPIGRPCPort),
					TargetPort: intstr.FromInt(appSidecarAPIGRPCPort),
					Name:       appSidecarAPIGRPCPortName,
				},
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(appSidecarInternalGRPCPort),
					TargetPort: intstr.FromInt(appSidecarInternalGRPCPort),
					Name:       appSidecarInternalGRPCPortName,
				},
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(metricsPort),
					TargetPort: intstr.FromInt(metricsPort),
					Name:       appSidecarMetricsPortName,
				},
			},
		},
	}
}

func (h *AppHandler) getAppID(wrapper ObjectWrapper) string {
	annotations := wrapper.GetTemplateAnnotations()
	if val, ok := annotations[appIDAnnotationKey]; ok && val != "" {
		return val
	}
	return ""
}

func (h *AppHandler) isAnnotatedForApp(wrapper ObjectWrapper) bool {
	annotations := wrapper.GetTemplateAnnotations()
	enabled, ok := annotations[appEnabledAnnotationKey]
	if !ok {
		return false
	}
	switch strings.ToLower(enabled) {
	case "y", "yes", "true", "on", "1":
		return true
	default:
		return false
	}
}

func (h *AppHandler) getEnableMetrics(wrapper ObjectWrapper) bool {
	annotations := wrapper.GetTemplateAnnotations()
	enableMetrics := defaultMetricsEnabled
	if val, ok := annotations[appEnableMetricsKey]; ok {
		if v, err := strconv.ParseBool(val); err == nil {
			enableMetrics = v
		}
	}
	return enableMetrics
}

func (h *AppHandler) getMetricsPort(wrapper ObjectWrapper) int {
	annotations := wrapper.GetTemplateAnnotations()
	metricsPort := defaultMetricsPort
	if val, ok := annotations[appMetricsPortKey]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			metricsPort = v
		}
	}
	return metricsPort
}
