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

package controller

import (
	"context"
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	configv1 "github.com/Nivesh00/config-keys-operator.git/api/v1"
)

// EnvKeyMonitorReconciler reconciles a EnvKeyMonitor object
type EnvKeyMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.core.nvsh-ram.io,resources=envkeymonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.core.nvsh-ram.io,resources=envkeymonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.core.nvsh-ram.io,resources=envkeymonitors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EnvKeyMonitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *EnvKeyMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// TODO(user): your logic here

	// Get new EnvKeyMonitor
	var envKeyMonitor configv1.EnvKeyMonitor
	if err := r.Get(ctx, req.NamespacedName, &envKeyMonitor); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("EnvKeyMonitor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get EnvKeyMonitor")
		return ctrl.Result{}, err
	}

	// Check for duplicate keys in current obj
	log.Info("Checking for duplicates in current EnvKeyMonitor object", "EnvKeyMonitor object name", envKeyMonitor.Name)
	if err := CheckDuplicateKeysInSingleObj(&envKeyMonitor.Spec.Keys); err != nil {
		log.Error(err, "Rejecting creation of new EnvKeyMonitor object...")
		return ctrl.Result{}, err
	}

	// Get all EnvKeyMonitor objs in namespace
	log.Info("Getting all EnvKeyMonitor objects in current namespace", "Namespace", envKeyMonitor.Namespace)
	var envKeyMonitorList configv1.EnvKeyMonitorList
	if err := r.List(ctx, &envKeyMonitorList, client.InNamespace(envKeyMonitor.Namespace)); err != nil {
		log.Info("Cannot get EnvKeyMonitor objects in current namespace")
		return ctrl.Result{}, err
	}

	// Check for duplicate keys in all EnvKeyMonitor objects in current namespace
	log.Info("Checking for duplicates in all EnvKeyMonitor objects in current namespace", "Namespace", envKeyMonitor.Namespace)
	if err := CheckDuplicateKeysInNamespace(&envKeyMonitor.Spec.Keys, &envKeyMonitorList); err != nil {
		log.Error(err, "Rejecting creation of new EnvKeyMonitor object...")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Check if there are duplicates in path .spec.keys[] for a single EnvKeyMonitor obj
func CheckDuplicateKeysInSingleObj(keysList *[]string) error {

	var accumulator []*string
	for _, key := range *keysList {
		// If duplicate is present, return an error
		if slices.Contains(accumulator, &key) {
			return fmt.Errorf(
				"Cannot create EnvKeyMonitor due to duplicates found in .spec.keys[]\n"+
					"EnvKeyMonitor cannot have duplicates in the 'keys' list Duplicate key found is %s",
				key,
			)
		}
		// Add key to accumulator
		accumulator = append(accumulator, &key)
	}
	return nil
}

// Check if there are duplicates in path .spec.keys[] for all EnvKeyMonitor obj in current namespace
func CheckDuplicateKeysInNamespace(newKeysList *[]string, envKeyMonitorList *configv1.EnvKeyMonitorList) error {

	// Get all keys currently in namespace
	var allKeysList []*string
	for _, envKeyMonitor := range envKeyMonitorList.Items {
		for _, key := range envKeyMonitor.Spec.Keys {
			allKeysList = append(allKeysList, &key)
		}
	}

	for _, key := range *newKeysList {
		// If duplicate is present, return an error
		if slices.Contains(allKeysList, &key) {
			return fmt.Errorf(
				"Cannot create EnvKeyMonitor due to duplicates found in .spec.keys[]\n"+
					"EnvKeyMonitor cannot have duplicates in the 'keys' list\n"+
					"An EnvKeyMonitor object already exists with key '%s' in target namespace",
				key,
			)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvKeyMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1.EnvKeyMonitor{}).
		Named("envkeymonitor").
		Complete(r)
}
