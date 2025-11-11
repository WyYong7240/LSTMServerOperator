/*
Copyright 2025 wuyong7240.

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

package controller

import (
	"context"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	lstmappsv1 "github.com/WyYong7240/LSTMServiceOperator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// LSTMPredictAppReconciler reconciles a LSTMPredictApp object
type LSTMPredictAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var CounterReconcileLSTMPredictApp int64

// 通用的重新排队的时间间隔
const GenericRequeueDuration = 1 * time.Minute

// +kubebuilder:rbac:groups=lstmapps.wuyong7240.com,resources=lstmpredictapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lstmapps.wuyong7240.com,resources=lstmpredictapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=lstmapps.wuyong7240.com,resources=lstmpredictapps/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LSTMPredictApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *LSTMPredictAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// 如果同时创建三个CRD资源，3个事件会被同时处理，防止日志混乱，加等待
	<-time.NewTicker(100 * time.Millisecond).C
	log := log.FromContext(ctx)

	// 用于统计Reconcile被调用了多少次
	CounterReconcileLSTMPredictApp += 1
	log.Info("Start LSTMPredictApp Reconcile", "number", CounterReconcileLSTMPredictApp)

	// 从上下文中获取CRD对象
	app := &lstmappsv1.LSTMPredictApp{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		// 如果是没找到，不用管
		if errors.IsNotFound(err) {
			log.Info("LSTMPredictApp not found.")
			return ctrl.Result{}, nil
		}
		// 如果不是没找到，那就要重新排队
		log.Error(err, "Failed to get the LSTMPredictApp, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 调谐子资源
	var result ctrl.Result
	var err error

	// 首先调谐Deployment，作为LSTM预测应用的后端应用
	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Deployment.")
		return result, err
	}

	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Service.")
		return result, err
	}

	log.Info("All resources have been reconciled.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LSTMPredictAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("Setup")
	return ctrl.NewControllerManagedBy(mgr).
		// 监听CR自定义资源的创建删除与更新
		For(&lstmappsv1.LSTMPredictApp{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				// 一旦创建该类型的CR，立即触发Reconcile，不论什么情况
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The LSTMPredictApp has been deleted.", "Name", event.Object.GetName())
				return false
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				// 只有当ResourceVersion不同，并且CR的Spec发生变化时，才会触发Reconcile
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				oldSpec := event.ObjectOld.(*lstmappsv1.LSTMPredictApp).Spec
				newSpec := event.ObjectNew.(*lstmappsv1.LSTMPredictApp).Spec

				return !reflect.DeepEqual(oldSpec, newSpec)
			},
		})).
		// 监听因CR资源而产生的Deployment资源
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			// Deployment是由Reconcile控制器自己创建的，无需响应
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The LSTMPredictApp Deployment has been deleted.", "Name", event.Object.GetName())
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				oldSpec := event.ObjectOld.(*appsv1.Deployment).Spec
				newSpec := event.ObjectNew.(*appsv1.Deployment).Spec

				return !reflect.DeepEqual(oldSpec, newSpec)
			},
		})).
		// 监听因CR资源而产生的Service资源
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			//Service是由Reconcile控制器自己创建的，无需响应
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The LSTMPredictApp Service has been deleted.", "Name", event.Object.GetName())
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				oldSpec := event.ObjectOld.(*corev1.Service).Spec
				newSpec := event.ObjectNew.(*corev1.Service).Spec

				return !reflect.DeepEqual(oldSpec, newSpec)
			},
		})).
		Named("lstmpredictapp").
		Complete(r)
}
