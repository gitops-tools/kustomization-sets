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

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/runtime/patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	logger.Info("kustomization set loaded")

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

	existingEntries := sets.New[sourcev1alpha1.ResourceRef]()
	if kustomizationSet.Status.Inventory != nil {
		existingEntries.Insert(kustomizationSet.Status.Inventory.Entries...)
	}

	entries := sets.New[sourcev1alpha1.ResourceRef]()
	for _, kustomization := range kustomizations {
		objMeta, err := object.RuntimeToObjMeta(&kustomization)
		if err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}
		ref := sourcev1alpha1.ResourceRef{
			ID:      objMeta.String(),
			Version: kustomization.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		}
		entries.Insert(ref)

		if existingEntries.Has(ref) {
			existing := &kustomizev1.Kustomization{}
			if err := r.Client.Get(ctx, types.NamespacedName{Name: kustomization.Name, Namespace: kustomization.Namespace}, existing); err != nil {
				return nil, fmt.Errorf("failed to load existing Kustomization: %w", err)
			}
			patchHelper, err := patch.NewHelper(existing, r.Client)
			if err != nil {
				return nil, fmt.Errorf("failed to create patch helper for Kustomization: %w", err)
			}
			existing.Spec = kustomization.Spec
			if err := patchHelper.Patch(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to update Kustomization: %w", err)
			}
			continue
		}

		controllerutil.SetControllerReference(kustomizationSet, &kustomization, r.Scheme)

		if err := r.Client.Create(ctx, &kustomization); err != nil {
			return nil, fmt.Errorf("failed to create Kustomization: %w", err)
		}
	}

	if kustomizationSet.Status.Inventory == nil {
		return &sourcev1alpha1.ResourceInventory{Entries: entries.SortedList(func(x, y sourcev1alpha1.ResourceRef) bool {
			return x.ID < y.ID
		})}, nil

	}
	kustomizationsToRemove := existingEntries.Difference(entries)
	if err := r.removeResourceRefs(ctx, kustomizationsToRemove.List()); err != nil {
		return nil, err
	}

	return &sourcev1alpha1.ResourceInventory{Entries: entries.SortedList(func(x, y sourcev1alpha1.ResourceRef) bool {
		return x.ID < y.ID
	})}, nil

}

func (r *KustomizationSetReconciler) removeResourceRefs(ctx context.Context, deletions []sourcev1alpha1.ResourceRef) error {
	for _, v := range deletions {
		objMeta, err := object.ParseObjMetadata(v.ID)
		if err != nil {
			return fmt.Errorf("failed to parse object ID %s for deletion: %w", v.ID, err)
		}
		k := kustomizev1.Kustomization{
			ObjectMeta: metav1.ObjectMeta{
				Name:      objMeta.Name,
				Namespace: objMeta.Namespace,
			},
		}
		if err := r.Client.Delete(ctx, &k); err != nil {
			return fmt.Errorf("failed to delete %v: %w", k, err)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizationSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1alpha1.KustomizationSet{}).
		Complete(r)
}
