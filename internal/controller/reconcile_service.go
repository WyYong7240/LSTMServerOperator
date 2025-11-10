package controller

import (
	"context"
	"fmt"

	lstmappsv1 "github.com/WyYong7240/LSTMServiceOperator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *LSTMPredictAppReconciler) reconcileService(ctx context.Context, app *lstmappsv1.LSTMPredictApp) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 先根据LSTMPredictApp中的Namespace和Name信息查询对应的Service是否存在
	var svc = &corev1.Service{}
	// types.NamespaceedName用于唯一标识Kubernetes集群中的资源,svc是一个指针，如果Get方法成功执行，这个指针指向从API服务器中获取的Service对象
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, svc)

	// 没有错误发生时，更新状态
	if err == nil {
		log.Info("The Service has already exist.")

		// 属性更新,逐个属性判断是否有变更,如果有变更再更新现存资源
		var isChanged bool = false
		svc.Spec.Ports[0].TargetPort = intstr.FromInt32(app.Spec.ContainerPort)
		if app.Spec.ServicePort != svc.Spec.Ports[0].Port {
			svc.Spec.Ports[0].Port = app.Spec.ServicePort
			isChanged = true
		}
		if app.Spec.ServiceType != svc.Spec.Type {
			svc.Spec.Type = app.Spec.ServiceType
			isChanged = true
		}
		if isChanged {
			if err = r.Update(ctx, svc); err != nil {
				log.Error(err, "Failed to Update Service, will requeue, after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			log.Info("The LSTMPredictApp ServiceSpec has been updated.")
		}

		// 状态更新
		var serviceEndpoint string
		switch svc.Spec.Type {
		case corev1.ServiceTypeClusterIP:
			serviceEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port)
		case corev1.ServiceTypeNodePort:
			serviceEndpoint = fmt.Sprintf("%s:%d", svc.Spec.ClusterIP, svc.Spec.Ports[0].NodePort)
		}
		if serviceEndpoint != app.Status.ServiceEndPoint {
			app.Status.ServiceEndPoint = serviceEndpoint
			// 如果不同，需要更新LSTMPredictApp的状态
			// 调用r.Status().Update更新LSTMPredictApp资源的状态
			if err = r.Status().Update(ctx, app); err != nil {
				log.Error(err, "Failed to update LSTMPredictApp status")
				// 返回一个带有重新排队时间的结果和错误，表示需要在一段时间后重试
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			log.Info("The LSTMPredictApp ServiceStatus has been updated.")
		}

		log.Info("The LSTMPredictApp ServiceStatus and ServiceSpec has not been changed.")
		return ctrl.Result{}, nil
	}

	// 如果不是NotFound的错误，即发生了其他错误，结束本轮调谐，一段时间后重试
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Service, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 根据LSTMPredictApp资源实例信息来构造Service实例
	newService := &corev1.Service{}
	newService.SetName(app.Name)
	newService.SetNamespace(app.Namespace)
	newService.SetLabels(app.Labels)

	newService.Spec = corev1.ServiceSpec{
		Type:     app.Spec.ServiceType,
		Selector: map[string]string{"app": app.Name},
		Ports: []corev1.ServicePort{
			{
				Port:       app.Spec.ServicePort,
				TargetPort: intstr.FromInt32(app.Spec.ContainerPort),
			},
		},
	}

	// 用于建立App里擦同与Service之间的父子关系：Kubernetes通过owner Reference实现级联删除，当LSTMPredictApp被删除时，Kubernetes
	// 会自动删除它创建的Service; r.scheme用来识别资源类型的Scheme，确保类型正确
	if err := ctrl.SetControllerReference(app, newService, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 在集群中创建Service：调用客户端的Create方法，将newDp提交到Kubernetes API Server
	if err := r.Create(ctx, newService); err != nil {
		log.Error(err, "Failed to create Service, will requeue, after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	log.Info("The Service has been created.")
	return ctrl.Result{}, nil
}
