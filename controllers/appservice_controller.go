/*
Copyright 2023 hczhang.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "github.com/best/k8s-opdemo/api/v1beta1"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.thislab.cn,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.thislab.cn,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.thislab.cn,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 首先获取 AppService 实例
	var appService appv1beta1.AppService
	err := r.Client.Get(ctx, req.NamespacedName, &appService)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("fetch appservice objects", "appservice", appService)

	// 得到 AppService 后去创建对应的 Deployment 和 Service（观察当前状态和期望状态进行对比）

	// 调谐：获取到当前的状态和期望的状态进行对比
	// Create or update deployment
	var deploy appsv1.Deployment
	deploy.Name = appService.Name
	deploy.Namespace = appService.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		// 调谐在这个函数中实现
		MutateDeployment(&appService, &deploy)
		return controllerutil.SetControllerReference(&appService, &deploy, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Create or update deployment", "result", or)

	// Create or update service
	var svc corev1.Service
	svc.Name = appService.Name
	svc.Namespace = appService.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		MutateService(&appService, &svc)
		return controllerutil.SetControllerReference(&appService, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Create or update service", "result", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
