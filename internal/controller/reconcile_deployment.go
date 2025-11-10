package controller

import (
	"context"

	lstmappsv1 "github.com/WyYong7240/LSTMServiceOperator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *LSTMPredictAppReconciler) reconcileDeployment(ctx context.Context, app *lstmappsv1.LSTMPredictApp) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 先根据LSTMPredictApp中的Namespace和Name信息查询对应的Deployment是否存在
	var dp = &appsv1.Deployment{}
	// types.NamespaceedName用于唯一标识Kubernetes集群中的资源,dp是一个指针，如果Get方法成功执行，这个指针指向从API服务器中获取的Deployment对象
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)

	// 没有错误发生，先判定Deployment属性是否与CR定义的一致，不一致的话改为一致；再更新对应的状态
	if err == nil {
		log.Info("The Deployment has already exist.")

		// 属性更新
		var isChanged bool = false
		dp.Spec.Template.Spec.Containers[0].Image = app.Spec.AppImage
		dp.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = app.Spec.ContainerPort
		if *app.Spec.BackendAppReplicas != *dp.Spec.Replicas {
			dp.Spec.Replicas = app.Spec.BackendAppReplicas
			isChanged = true
		}
		if !isEmptyResourceRequirements(app.Spec.ResourcesLimit) {
			dp.Spec.Template.Spec.Containers[0].Resources = app.Spec.ResourcesLimit
		}
		if isChanged {
			if err = r.Update(ctx, dp); err != nil {
				log.Error(err, "Failed to Update Deployment, will requeue, after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			log.Info("LSTMPredictApp Deployment Update Success!")
		}

		// 状态更新
		// 更新当前已经Ready的副本数量
		app.Status.ReadyReplicas = dp.Status.ReadyReplicas
		// 如果副本数量达到了要求的数量，则CR的状态中Phase变为running，否则是Pending
		if dp.Status.ReadyReplicas == *app.Spec.BackendAppReplicas {
			app.Status.Phase = "Running"
		} else {
			app.Status.Phase = "Pending"
		}
		// 每次更新都会触发Reconcile，所以在这里更新最近一次更新时间
		app.Status.LastUpdateTime = metav1.Now()

		// 调用r.Status().Update更新LSTMPredictApp资源的状态
		if err = r.Status().Update(ctx, app); err != nil {
			log.Error(err, "Failed to update LSTMPredictApp status.")
			// 返回一个带有重新排队时间的结果和错误，表示需要在一段时间后重试
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("The LSTMPredictApp status has been updated.")
		return ctrl.Result{}, nil
	}

	// 如果不是NotFound的错误，即发生了其他错误，结束本轮调谐，一段时间后重试
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Deployment, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 根据LSTMPredictApp资源实例信息来构造Deployment实例
	newDp := &appsv1.Deployment{}
	newDp.SetName(app.Name)
	newDp.SetNamespace(app.Namespace)
	newDp.SetLabels(app.Labels)

	// 对app.Spec.BackendAppReplicas为空时赋默认值处理值

	newDp.Spec = appsv1.DeploymentSpec{
		Replicas: app.Spec.BackendAppReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": app.Name},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": app.Name},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "lstm-predict-app",
						Image: app.Spec.AppImage,
						Ports: []corev1.ContainerPort{
							{ContainerPort: app.Spec.ContainerPort},
						},
					},
				},
			},
		},
	}

	// 对app.Spec.ResoucesLimit为空进行处理，如果为空，不对Pod的资源限制做出定义
	if !isEmptyResourceRequirements(app.Spec.ResourcesLimit) {
		newDp.Spec.Template.Spec.Containers[0].Resources = app.Spec.ResourcesLimit
	}

	// 用于建立App里擦同与Deployment之间的父子关系：Kubernetes通过owner Reference实现级联删除，当LSTMPredictApp被删除时，Kubernetes
	// 会自动删除它创建的Deployment; r.scheme用来识别资源类型的Scheme，确保类型正确
	if err := ctrl.SetControllerReference(app, newDp, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 在集群中创建Deployment：调用客户端的Create方法，将newDp提交到Kubernetes API Server
	if err := r.Create(ctx, newDp); err != nil {
		log.Error(err, "Failed to create Deployment, will requeue, after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	log.Info("The Deployment has been created.")
	return ctrl.Result{}, nil
}

func isEmptyResourceRequirements(r corev1.ResourceRequirements) bool {
	return len(r.Limits) == 0 && len(r.Requests) == 0
}
