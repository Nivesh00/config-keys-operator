/*
Copyright 2025.

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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	configv1 "github.com/Nivesh00/config-keys-operator.git/api/v1"
)

// nolint:unused
// log is for logging in this package.
var configmaplog = logf.Log.WithName("configmap-resource")

// SetupConfigMapWebhookWithManager registers the webhook for ConfigMap in the manager.
func SetupConfigMapWebhookWithManager(mgr ctrl.Manager) error {

	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.ConfigMap{}).
		WithValidator(&ConfigMapCustomValidator{
			mgr.GetClient(),
		}).
		WithValidatorCustomPath("/env-keys-validation").
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/env-keys-validation,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=configmaps,verbs=create;update,versions=v1,name=vconfigmap-v1.kb.io,admissionReviewVersions=v1

// ConfigMapCustomValidator struct is responsible for validating the ConfigMap resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ConfigMapCustomValidator struct {

	// TODO(user): Add more fields as needed for validation
	// Used to query k8s
	client.Client
}

var _ webhook.CustomValidator = &ConfigMapCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ConfigMap.
func (v *ConfigMapCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("expected a ConfigMap object but got %T", obj)
	}
	configmaplog.Info("Validation for ConfigMap upon creation", "name", configmap.GetName())

	// (user): fill in your validation logic upon object creation.

	// Get list of existing EnvKeyMonitors
	var envKeyMonitorList configv1.EnvKeyMonitorList
	if err := v.List(ctx, &envKeyMonitorList, client.InNamespace(configmap.Namespace)); err != nil {
		configmaplog.Info(err.Error() + " Cannot get EnvKeyMonitor CRDs in namespace. Rejecting configmap creation")
		return nil, fmt.Errorf("failed to list configmaps: %v", err)
	}

	// Get all forbidden keys
	forbiddenKeysList := getEnvKevMonitorKeys(&envKeyMonitorList)

	configmaplog.Info("Configmap which contain the following keys are not allowed in the current namespace",
		"namespace",
		configmap.Namespace,
		"forbidden keys",
		strings.Join(*forbiddenKeysList, ", "),
	)

	// Check if configmap contains a forbidden key
	configmaplog.Info("Checking if configmap contains forbidden keys...")
	if err := checkConfigmapKeys(forbiddenKeysList, configmap); err != nil {
		configmaplog.Info(err.Error() + " rejecting configmap...")
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ConfigMap.
func (v *ConfigMapCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	configmap, ok := newObj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("expected a ConfigMap object for the newObj but got %T", newObj)
	}
	configmaplog.Info("Validation for ConfigMap upon update", "name", configmap.GetName())

	// TODO(user): fill in your validation logic upon object update.

	// Get list of existing EnvKeyMonitors
	var envKeyMonitorList configv1.EnvKeyMonitorList
	if err := v.List(ctx, &envKeyMonitorList, client.InNamespace(configmap.Namespace)); err != nil {
		configmaplog.Info(err.Error() + " Cannot get EnvKeyMonitor CRDs in namespace. Rejecting configmap creation")
		return nil, fmt.Errorf("failed to list configmaps: %v", err)
	}

	// Get all forbidden keys
	forbiddenKeysList := getEnvKevMonitorKeys(&envKeyMonitorList)

	configmaplog.Info("Configmap which contain the following keys are not allowed in the current namespace",
		"namespace",
		configmap.Namespace,
		"forbidden keys",
		strings.Join(*forbiddenKeysList, ", "),
	)

	// Check if configmap contains a forbidden key
	configmaplog.Info("Checking if configmap contains forbidden keys...")
	if err := checkConfigmapKeys(forbiddenKeysList, configmap); err != nil {
		configmaplog.Info(err.Error() + " rejecting configmap...")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ConfigMap.
func (v *ConfigMapCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("expected a ConfigMap object but got %T", obj)
	}
	configmaplog.Info("Validation for ConfigMap upon deletion", "name", configmap.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// Get a list of all EnvKeyMonitor keys in namespace
func getEnvKevMonitorKeys(envKeyMonitorList *configv1.EnvKeyMonitorList) *[]string {

	var allKeys []string
	for _, envKeyMonitor := range envKeyMonitorList.Items {
		for _, key := range envKeyMonitor.Spec.Keys {
			allKeys = append(allKeys, key)
		}
	}
	return &allKeys
}

// Check if EnvKeyMonitor CRD in current namespace contain a key
func checkConfigmapKeys(forbiddenKeysList *[]string, configmap *corev1.ConfigMap) error {

	for _, forbiddenKey := range *forbiddenKeysList {
		if _, keyExists := configmap.Data[forbiddenKey]; keyExists {
			return fmt.Errorf(
				"Configmap contains forbidden key and is therefore invalid. "+
					"Forbidden key is '%s'",
				forbiddenKey,
			)
		}
	}
	return nil
}
