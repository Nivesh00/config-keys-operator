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
	"slices"

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
var envKeyMonitorLog = logf.Log.WithName("envkeymonitor-resource")

// SetupEnvKeyMonitorWebhookWithManager registers the webhook for EnvKeyMonitor in the manager.
func SetupEnvKeyMonitorWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&configv1.EnvKeyMonitor{}).
		WithValidator(&EnvKeyMonitorCustomValidator{
			mgr.GetClient(),
		}).
		WithValidatorCustomPath("/envkeymonitor-validate").
		WithDefaulter(&EnvKeyMonitorCustomDefaulter{
			mgr.GetClient(),
		}).
		WithDefaulterCustomPath("/envkeymonitor-mutate").
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/envkeymonitor-mutate,mutating=true,failurePolicy=fail,sideEffects=None,groups=config.core.nvsh-ram.io,resources=envkeymonitors,verbs=create;update,versions=v1,name=menvkeymonitor-v1.kb.io,admissionReviewVersions=v1

// EnvKeyMonitorCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind EnvKeyMonitor when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type EnvKeyMonitorCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	client.Client
}

var _ webhook.CustomDefaulter = &EnvKeyMonitorCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind EnvKeyMonitor.
func (d *EnvKeyMonitorCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	envKeyMonitor, ok := obj.(*configv1.EnvKeyMonitor)

	if !ok {
		return fmt.Errorf("expected an EnvKeyMonitor object but got %T", obj)
	}
	envKeyMonitorLog.Info("Defaulting for EnvKeyMonitor", "name", envKeyMonitor.GetName())

	// TODO(user): fill in your defaulting logic.

	// Remove duplicates in new object
	envKeyMonitor.Spec.Keys = d.RemoveDuplicatesInObject(envKeyMonitor)

	// Remove duplicates in new object if EnvKeyMonitor in current namespace contains it
	newKeys, err := d.RemoveDuplicatesInNamespace(&ctx, envKeyMonitor)
	if err != nil {
		envKeyMonitorLog.Error(err, "Cannot process new EnvKeyObject object. Object is invalid. Rejecting new object...")
		return fmt.Errorf("An error occured while processing new EnvKeyMonitor object. Cannot create new object")
	}
	envKeyMonitor.Spec.Keys = newKeys

	return nil
}

// Check new object only to remove duplicates found in .spec.keys[]
func (d *EnvKeyMonitorCustomDefaulter) RemoveDuplicatesInObject(envKeyMonitor *configv1.EnvKeyMonitor) []string {

	var newKeysList []string
	for _, key := range envKeyMonitor.Spec.Keys {
		if slices.Contains(newKeysList, key) {
			envKeyMonitorLog.Info(
				"Object contains duplicate in '.spec.keys[]'. Removing duplicated key...",
				"name",
				envKeyMonitor.GetName(),
				"namespace",
				envKeyMonitor.GetNamespace(),
				"duplicate_key",
				key,
			)
			continue
		}
		newKeysList = append(newKeysList, key)
	}
	return newKeysList
}

// Check new object and all EnvKeyMonitor CRDs in namespace to remove duplicates found in .spec.keys[]
func (d *EnvKeyMonitorCustomDefaulter) RemoveDuplicatesInNamespace(ctx *context.Context, envKeyMonitor *configv1.EnvKeyMonitor) ([]string, error) {

	// Get list of EnvKeyMonitor
	var envKeyMonitorList configv1.EnvKeyMonitorList
	if err := d.List(*ctx, &envKeyMonitorList, client.InNamespace(envKeyMonitor.GetNamespace())); err != nil {
		envKeyMonitorLog.Info(
			"Failed to check for duplicates, cannot list EnvKeyMonitor CRDs in namespace",
			"namespace",
			envKeyMonitor.GetNamespace(),
		)
		return nil, err
	}

	// Get all .spec.keys[] from list of EnvKeyMonitor list
	var currentKeysList []string
	for _, item := range envKeyMonitorList.Items {
		if item.GetName() == envKeyMonitor.GetName() {
			continue
		}
		currentKeysList = append(currentKeysList, item.Spec.Keys...)
	}

	// Check and remove duplicates in new
	var newKeysList []string
	for _, key := range envKeyMonitor.Spec.Keys {
		if slices.Contains(currentKeysList, key) {

			envKeyMonitorLog.Info(
				"Object contains duplicate in '.spec.keys[]', another object in the current namespace "+
					"already contains this key. Removing duplicated key...",
				"name",
				envKeyMonitor.GetName(),
				"namespace",
				envKeyMonitor.GetNamespace(),
				"duplicate_key",
				key,
			)
			continue
		}
		newKeysList = append(newKeysList, key)
	}

	return newKeysList, nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/envkeymonitor-validate,mutating=false,failurePolicy=fail,sideEffects=None,groups=config.core.nvsh-ram.io,resources=envkeymonitors,verbs=create;update,versions=v1,name=venvkeymonitor-v1.kb.io,admissionReviewVersions=v1

// EnvKeyMonitorCustomValidator struct is responsible for validating the EnvKeyMonitor resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type EnvKeyMonitorCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
	client.Client
}

var _ webhook.CustomValidator = &EnvKeyMonitorCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type EnvKeyMonitor.
func (v *EnvKeyMonitorCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	envKeyMonitor, ok := obj.(*configv1.EnvKeyMonitor)
	if !ok {
		return nil, fmt.Errorf("expected a EnvKeyMonitor object but got %T", obj)
	}
	envKeyMonitorLog.Info("Validation for EnvKeyMonitor upon creation", "name", envKeyMonitor.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	// Check for duplicates in object
	if err := v.CheckDuplicateKeysInObject(envKeyMonitor); err != nil {
		envKeyMonitorLog.Error(err, "Cannot create new object")
		return nil, err
	}
	// Check for duplicates in namespace
	if err := v.CheckDuplicateKeysInNamespace(&ctx, envKeyMonitor); err != nil {
		envKeyMonitorLog.Error(err, "Cannot create new object")
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type EnvKeyMonitor.
func (v *EnvKeyMonitorCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	envKeyMonitor, ok := newObj.(*configv1.EnvKeyMonitor)
	if !ok {
		return nil, fmt.Errorf("expected a EnvKeyMonitor object for the newObj but got %T", newObj)
	}
	envKeyMonitorLog.Info("Validation for EnvKeyMonitor upon update", "name", envKeyMonitor.GetName())

	// TODO(user): fill in your validation logic upon object update.

	// Check for duplicates in object
	if err := v.CheckDuplicateKeysInObject(envKeyMonitor); err != nil {
		envKeyMonitorLog.Error(err, "Cannot create new object")
		return nil, err
	}
	// Check for duplicates in namespace
	if err := v.CheckDuplicateKeysInNamespace(&ctx, envKeyMonitor); err != nil {
		envKeyMonitorLog.Error(err, "Cannot create new object")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type EnvKeyMonitor.
func (v *EnvKeyMonitorCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	envKeyMonitor, ok := obj.(*configv1.EnvKeyMonitor)
	if !ok {
		return nil, fmt.Errorf("expected a EnvKeyMonitor object but got %T", obj)
	}
	envKeyMonitorLog.Info("Validation for EnvKeyMonitor upon deletion", "name", envKeyMonitor.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// Check if there are duplicates in current object
func (v *EnvKeyMonitorCustomValidator) CheckDuplicateKeysInObject(envKeyMonitor *configv1.EnvKeyMonitor) error {

	var newKeysList []string
	for _, key := range envKeyMonitor.Spec.Keys {
		if slices.Contains(newKeysList, key) {

			envKeyMonitorLog.Info(
				"Duplicate keys found in object during validation",
				"name",
				envKeyMonitor.GetName(),
				"namespace",
				envKeyMonitor.GetNamespace(),
				"duplicate_key",
				key,
			)

			return fmt.Errorf(
				"Duplicate keys found in EnvKeyMonitor object during validation. "+
					"Key %s is listed more than once in EnvKeyMonitor object %s",
				key,
				envKeyMonitor.GetName(),
			)
		}
		newKeysList = append(newKeysList, key)
	}

	return nil
}

// Check if there are duplicates in current namespace
func (v *EnvKeyMonitorCustomValidator) CheckDuplicateKeysInNamespace(ctx *context.Context, envKeyMonitor *configv1.EnvKeyMonitor) error {

	// Get list of EnvKeyMonitor
	var envKeyMonitorList configv1.EnvKeyMonitorList
	if err := v.List(*ctx, &envKeyMonitorList, client.InNamespace(envKeyMonitor.GetNamespace())); err != nil {

		envKeyMonitorLog.Info(
			"Failed to check for duplicates, cannot list objects CRDs in namespace",
			"name",
			envKeyMonitor.GetName(),
			"namespace",
			envKeyMonitor.GetNamespace(),
		)

		return fmt.Errorf("Cannot list object ")
	}

	// Get all .spec.keys[] from list of EnvKeyMonitor list
	var currentKeysList []string
	for _, item := range envKeyMonitorList.Items {
		if item.GetName() == envKeyMonitor.GetName() {
			continue
		}
		currentKeysList = append(currentKeysList, item.Spec.Keys...)
	}

	// Check and remove duplicates in new
	var newKeysList []string
	for _, key := range envKeyMonitor.Spec.Keys {
		if slices.Contains(currentKeysList, key) {

			envKeyMonitorLog.Info(
				"Object contains duplicate in '.spec.keys[]', another object in the current namespace"+
					"already contains this key. Removing duplicated key...",
				"name",
				envKeyMonitor.GetName(),
				"namespace",
				envKeyMonitor.GetNamespace(),
				"duplicate_key",
				key,
			)

			return fmt.Errorf(
				"Another object of kind %s in current namespace with key %s "+
					"already exists in namespace %s",
				envKeyMonitor.GetObjectKind(),
				key,
				envKeyMonitor.GetNamespace(),
			)
		}

		newKeysList = append(newKeysList, key)
	}

	return nil
}
