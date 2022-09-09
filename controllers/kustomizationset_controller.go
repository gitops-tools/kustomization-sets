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
	"sigs.k8s.io/cli-utils/pkg/object"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sourcev1alpha1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators"
	"github.com/gitops-tools/pkg/sets"
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

	if !kustomizationSet.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	inventory, err := r.reconcileResources(ctx, &kustomizationSet)
	if err != nil {
		return ctrl.Result{}, err
	}
	if inventory != nil {
		kustomizationSet.Status.Inventory = inventory
		if err := r.Status().Update(ctx, &kustomizationSet); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KustomizationSetReconciler) reconcileResources(ctx context.Context, kustomizationSet *sourcev1alpha1.KustomizationSet) (*sourcev1alpha1.ResourceInventory, error) {
	kustomizations, err := reconciler.GenerateKustomizations(ctx, kustomizationSet, r.Generators)
	if err != nil {
		return nil, err
	}
	entries := sets.New[sourcev1alpha1.ResourceRef]()
	// TODO: This should check for existing resources and update rather than
	// create.
	for _, kustomization := range kustomizations {
		if err := controllerutil.SetOwnerReference(kustomizationSet, &kustomization, r.Client.Scheme()); err != nil {
			return nil, fmt.Errorf("failed to set owner for Kustomization: %w", err)
		}
		if err := r.Client.Create(ctx, &kustomization); err != nil {
			return nil, fmt.Errorf("failed to create Kustomization: %w", err)
		}
		objMeta, err := object.RuntimeToObjMeta(&kustomization)
		if err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}
		entries.Insert(sourcev1alpha1.ResourceRef{
			ID:      objMeta.String(),
			Version: kustomization.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		})
	}

	// TODO: This is the point to delete and remove existing resources.
	// if kustomizationSet.Status.Inventory != nil {
	// 	previouslyGenerated := sets.New(kustomizationSet.Status.Inventory.Entries...)
	// 	kustomizationsToRemove := previouslyGenerated.Difference(entries)
	// }

	return &sourcev1alpha1.ResourceInventory{Entries: entries.SortedList(func(x, y sourcev1alpha1.ResourceRef) bool {
		return x.ID < y.ID
	})}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizationSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1alpha1.KustomizationSet{}).
		Complete(r)
}
