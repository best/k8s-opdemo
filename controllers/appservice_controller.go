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
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/retry"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "github.com/best/k8s-opdemo/api/v1beta1"
)

var (
	oldSpecAnnotation = "app.thislab.cn/old-spec"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
	log := log.FromContext(ctx)

	// 首先获取 AppService 实例
	var appService appv1beta1.AppService
	err := r.Client.Get(ctx, req.NamespacedName, &appService)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, nil
		}

		// 说明已经被删除了，无需重新入队列
		return ctrl.Result{}, nil
	}

	log.Info("fetch appservice objects", "appservice", appService)

	// 当前对象已被标记为删除
	if appService.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// 如果不存在管理的资源，去创建
	deploy := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, deploy); err != nil && client.IgnoreNotFound(err) == nil {
		// 关联 Annotations
		data, err := json.Marshal(appService.Spec)
		if err != nil {
			return ctrl.Result{}, err
		}
		if appService.Annotations != nil {
			appService.Annotations[oldSpecAnnotation] = string(data)
		} else {
			appService.Annotations = map[string]string{
				oldSpecAnnotation: string(data),
			}
		}

		// 重新更新 AppService
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Client.Update(ctx, &appService)
		}); err != nil {
			return ctrl.Result{}, err
		}

		// 创建 Deployment
		newDeploy := NewDeploy(&appService)
		if err := r.Client.Create(ctx, newDeploy); err != nil {
			return ctrl.Result{}, err
		}

		// 创建 Service
		newService := NewService(&appService)
		if err := r.Client.Create(ctx, newService); err != nil {
			return ctrl.Result{}, err
		}

		// 创建成功
		return ctrl.Result{}, nil
	}

	// 如果存在关联的资源，判断是否去更新
	oldSpec := appv1beta1.AppServiceSpec{}
	if err := json.Unmarshal([]byte(appService.Annotations[oldSpecAnnotation]), &oldSpec); err != nil {
		return ctrl.Result{}, err
	}

	// 如果资源有变化，更新资源
	if !reflect.DeepEqual(oldSpec, appService.Spec) {
		// 更新 Deployment
		newDeploy := NewDeploy(&appService)
		oldDeploy := &appsv1.Deployment{}
		if err := r.Client.Get(ctx, req.NamespacedName, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}
		oldDeploy.Spec = newDeploy.Spec
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Client.Update(ctx, oldDeploy)
		}); err != nil {
			return ctrl.Result{}, err
		}

		// 更新 Service
		newService := NewService(&appService)
		oldService := &corev1.Service{}
		if err := r.Client.Get(ctx, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		oldService.Spec = newService.Spec
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Client.Update(ctx, oldService)
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Complete(r)
}
