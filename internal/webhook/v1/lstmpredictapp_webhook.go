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

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	lstmappsv1 "github.com/WyYong7240/LSTMServiceOperator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// nolint:unused
// log is for logging in this package.
var lstmpredictapplog = logf.Log.WithName("lstmpredictapp-resource")

// SetupLSTMPredictAppWebhookWithManager registers the webhook for LSTMPredictApp in the manager.
func SetupLSTMPredictAppWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&lstmappsv1.LSTMPredictApp{}).
		WithValidator(&LSTMPredictAppCustomValidator{
			MaxBackendAppReplicas: 10,
			MinBackendAppReplicas: 1,
			MaxPortID:             30000,
			AvailableServiceType:  []string{"ClusterIP", "NodePort"},
		}).
		WithDefaulter(&LSTMPredictAppCustomDefaulter{
			DefaultBackendAppReplicas: 1,
			DefaultServicePort:        8001,
			DefaultServiceType:        "ClusterIP",
			MinResourcesLimit: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-lstmapps-wuyong7240-com-v1-lstmpredictapp,mutating=true,failurePolicy=fail,sideEffects=None,groups=lstmapps.wuyong7240.com,resources=lstmpredictapps,verbs=create;update,versions=v1,name=mlstmpredictapp-v1.kb.io,admissionReviewVersions=v1

// LSTMPredictAppCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind LSTMPredictApp when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type LSTMPredictAppCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	DefaultBackendAppReplicas int32
	DefaultServicePort        int32
	DefaultServiceType        string
	MinResourcesLimit         corev1.ResourceRequirements
}

var _ webhook.CustomDefaulter = &LSTMPredictAppCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind LSTMPredictApp.
func (d *LSTMPredictAppCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	lstmpredictapp, ok := obj.(*lstmappsv1.LSTMPredictApp)

	if !ok {
		return fmt.Errorf("expected an LSTMPredictApp object but got %T", obj)
	}
	lstmpredictapplog.Info("Defaulting for LSTMPredictApp", "name", lstmpredictapp.GetName())

	// TODO(user): fill in your defaulting logic.
	// 后端应用副本数默认值注入
	if lstmpredictapp.Spec.BackendAppReplicas == nil {
		lstmpredictapp.Spec.BackendAppReplicas = new(int32)
		*lstmpredictapp.Spec.BackendAppReplicas = d.DefaultBackendAppReplicas
	}
	// 应用资源限制默认值注入
	if isEmptyResourceRequirements(lstmpredictapp.Spec.ResourcesLimit) {
		lstmpredictapp.Spec.ResourcesLimit = *d.MinResourcesLimit.DeepCopy()
	}
	// 服务类型默认值注入
	if lstmpredictapp.Spec.ServiceType == "" {
		lstmpredictapp.Spec.ServiceType = corev1.ServiceType(d.DefaultServiceType)
	}
	// 服务端口默认值注入
	if lstmpredictapp.Spec.ServicePort == 0 {
		lstmpredictapp.Spec.ServicePort = d.DefaultServicePort
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-lstmapps-wuyong7240-com-v1-lstmpredictapp,mutating=false,failurePolicy=fail,sideEffects=None,groups=lstmapps.wuyong7240.com,resources=lstmpredictapps,verbs=create;update,versions=v1,name=vlstmpredictapp-v1.kb.io,admissionReviewVersions=v1

// LSTMPredictAppCustomValidator struct is responsible for validating the LSTMPredictApp resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type LSTMPredictAppCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
	MaxBackendAppReplicas int32
	MinBackendAppReplicas int32
	MaxPortID             int32
	AvailableServiceType  []string
}

var _ webhook.CustomValidator = &LSTMPredictAppCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type LSTMPredictApp.
func (v *LSTMPredictAppCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	lstmpredictapp, ok := obj.(*lstmappsv1.LSTMPredictApp)
	if !ok {
		return nil, fmt.Errorf("expected a LSTMPredictApp object but got %T", obj)
	}
	lstmpredictapplog.Info("Validation for LSTMPredictApp upon creation", "name", lstmpredictapp.GetName())

	// TODO(user): fill in your validation logic upon object creation.
	if err := v.validateLSTMPredictAppSpec(lstmpredictapp); err != nil {
		return admission.Warnings{"LSTMPredictApp Webhook v1 Errors!"}, err
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type LSTMPredictApp.
func (v *LSTMPredictAppCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	lstmpredictapp, ok := newObj.(*lstmappsv1.LSTMPredictApp)
	if !ok {
		return nil, fmt.Errorf("expected a LSTMPredictApp object for the newObj but got %T", newObj)
	}
	lstmpredictapplog.Info("Validation for LSTMPredictApp upon update", "name", lstmpredictapp.GetName())

	// TODO(user): fill in your validation logic upon object update.
	if err := v.validateLSTMPredictAppSpec(lstmpredictapp); err != nil {
		return admission.Warnings{"LSTMPredictApp Webhook v1 Errors!"}, err
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type LSTMPredictApp.
func (v *LSTMPredictAppCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	lstmpredictapp, ok := obj.(*lstmappsv1.LSTMPredictApp)
	if !ok {
		return nil, fmt.Errorf("expected a LSTMPredictApp object but got %T", obj)
	}
	lstmpredictapplog.Info("Validation for LSTMPredictApp upon deletion", "name", lstmpredictapp.GetName())

	// TODO(user): fill in your validation logic upon object deletion.
	if err := v.validateLSTMPredictAppSpec(lstmpredictapp); err != nil {
		return admission.Warnings{"LSTMPredictApp Webhook v1 Errors!"}, err
	}
	return nil, nil
}

func isEmptyResourceRequirements(r corev1.ResourceRequirements) bool {
	return len(r.Limits) == 0 && len(r.Requests) == 0
}

func (v *LSTMPredictAppCustomValidator) validateLSTMPredictAppSpec(lstmpredictapp *lstmappsv1.LSTMPredictApp) error {
	// 校验副本数量，不小于设定的最小值
	if *lstmpredictapp.Spec.BackendAppReplicas < v.MinBackendAppReplicas {
		return fmt.Errorf("BackendAppReplicas can't < %d", v.MinBackendAppReplicas)
	}
	// 校验副本数量，不超过设定的最大值
	if *lstmpredictapp.Spec.BackendAppReplicas > v.MaxBackendAppReplicas {
		return fmt.Errorf("BackendAppReplicas can't > %d", v.MaxBackendAppReplicas)
	}
	// 校验服务端口号，须小于设定的最大端口号
	if lstmpredictapp.Spec.ServicePort >= v.MaxPortID {
		return fmt.Errorf("ServicePort is illeagle, need to < %d", v.MaxPortID)
	}

	// 校验服务类型，保证只在目前支持的服务类型中
	var isAvailable bool = false
	var currentAvailabelType string = "{"
	for _, serviceType := range v.AvailableServiceType {
		currentAvailabelType += serviceType + "/"
		if lstmpredictapp.Spec.ServiceType == corev1.ServiceType(serviceType) {
			isAvailable = true
		}
	}
	currentAvailabelType += "}"
	if !isAvailable {
		return fmt.Errorf("ServiceType is Unsupport, should be in %s", currentAvailabelType)
	}

	return nil
}
