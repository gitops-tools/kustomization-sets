/*
Copyright 2021.

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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sourcev1alpha1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler/generators"
)

// KustomizationSetReconciler reconciles a KustomizationSet object
type KustomizationSetReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Generators map[string]generators.Generator
}

//+kubebuilder:rbac:groups=source.gitops.solutions,resources=kustomizationsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.gitops.solutions,resources=kustomizationsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.gitops.solutions,resources=kustomizationsets/finalizers,verbs=update
//+kubebuilder:rbac:groups=kustomize.toolkit.fluxcd.io,resources=kustomizations,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KustomizationSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var kustomizationSet sourcev1alpha1.KustomizationSet
	if err := r.Client.Get(ctx, req.NamespacedName, &kustomizationSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("kustomization set loaded", "name", kustomizationSet.GetName())

	kustomizations, err := reconciler.GenerateKustomizations(ctx, &kustomizationSet, r.Generators)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, kustomization := range kustomizations {
		if err := controllerutil.SetOwnerReference(&kustomizationSet, &kustomization, r.Client.Scheme()); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set owner for Kustomization: %w", err)
		}
		if err := r.Client.Create(ctx, &kustomization); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create Kustomization: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizationSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1alpha1.KustomizationSet{}).
		Complete(r)
}
