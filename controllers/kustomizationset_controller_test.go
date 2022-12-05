/*
Copyright 2022.

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
	"path/filepath"
	"sort"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/cli-utils/pkg/object"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1alpha1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators/list"
	"github.com/google/go-cmp/cmp"
)

func TestReconciliation(t *testing.T) {
	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			"testdata/crds",
		},
	}
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatal(err)
	}

	if err := kustomizev1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatal(err)
	}
	if err := sourcev1alpha1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatal(err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatal(err)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatal(err)
	}
	reconciler := &KustomizationSetReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
		Generators: map[string]generators.Generator{
			"List": list.NewGenerator(),
		},
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		t.Fatal(err)
	}

	t.Run("reconciling creation of new resources", func(t *testing.T) {
		ctx := context.TODO()
		kz := newKustomizationSet()
		if err := k8sClient.Create(ctx, kz); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, kz)
		defer deleteAllKustomizations(t, k8sClient)

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(kz)})
		if err != nil {
			t.Fatal(err)
		}

		updated := &sourcev1alpha1.KustomizationSet{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kz), updated); err != nil {
			t.Fatal(err)
		}

		want := []runtime.Object{
			newKustomization("engineering-dev-demo", "default"),
			newKustomization("engineering-prod-demo", "default"),
			newKustomization("engineering-preprod-demo", "default"),
		}
		assertInventoryHasItems(t, updated, want...)
		assertKustomizationsExist(t, k8sClient, "default", "engineering-dev-demo", "engineering-prod-demo", "engineering-preprod-demo")
		assertKustomizationCondition(t, updated, meta.ReadyCondition, "3 kustomizations created")
	})

	t.Run("reconciling removal of resources", func(t *testing.T) {
		ctx := context.TODO()
		devKS := newKustomization("engineering-dev-demo", "default")
		kz := newKustomizationSet(func(ks *sourcev1alpha1.KustomizationSet) {
			ks.Spec.Generators = []sourcev1alpha1.KustomizationSetGenerator{
				{
					List: &sourcev1alpha1.ListGenerator{
						Elements: []apiextensionsv1.JSON{
							{Raw: []byte(`{"cluster": "engineering-prod"}`)},
							{Raw: []byte(`{"cluster": "engineering-preprod"}`)},
						},
					},
				},
			}
		})
		// TODO: create and cleanup
		if err := k8sClient.Create(ctx, kz); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, kz)
		if err := k8sClient.Create(ctx, devKS); err != nil {
			t.Fatal(err)
		}
		defer deleteAllKustomizations(t, k8sClient)

		objMeta, err := object.RuntimeToObjMeta(devKS)
		if err != nil {
			t.Fatal(err)
		}
		kz.Status.Inventory = &sourcev1alpha1.ResourceInventory{
			Entries: []sourcev1alpha1.ResourceRef{
				{
					ID:      objMeta.String(),
					Version: devKS.GetObjectKind().GroupVersionKind().GroupVersion().String(),
				},
			},
		}
		if err := k8sClient.Status().Update(ctx, kz); err != nil {
			t.Fatal(err)
		}

		_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(kz)})
		if err != nil {
			t.Fatal(err)
		}

		updated := &sourcev1alpha1.KustomizationSet{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kz), updated); err != nil {
			t.Fatal(err)
		}

		want := []runtime.Object{
			newKustomization("engineering-prod-demo", "default"),
			newKustomization("engineering-preprod-demo", "default"),
		}
		assertInventoryHasItems(t, updated, want...)
		assertResourceDoesNotExist(t, k8sClient, devKS)
	})

	t.Run("reconciling update of resources", func(t *testing.T) {
		ctx := context.TODO()
		devKS := newKustomization("engineering-dev-demo", "default")
		kz := newKustomizationSet(func(ks *sourcev1alpha1.KustomizationSet) {
			ks.Spec.Generators = []sourcev1alpha1.KustomizationSetGenerator{
				{
					List: &sourcev1alpha1.ListGenerator{
						Elements: []apiextensionsv1.JSON{
							{Raw: []byte(`{"cluster": "engineering-dev"}`)},
						},
					},
				},
			}
		})
		// TODO: create and cleanup
		if err := k8sClient.Create(ctx, kz); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, kz)
		if err := k8sClient.Create(ctx, devKS); err != nil {
			t.Fatal(err)
		}
		defer deleteAllKustomizations(t, k8sClient)

		objMeta, err := object.RuntimeToObjMeta(devKS)
		if err != nil {
			t.Fatal(err)
		}
		kz.Status.Inventory = &sourcev1alpha1.ResourceInventory{
			Entries: []sourcev1alpha1.ResourceRef{
				{
					ID:      objMeta.String(),
					Version: devKS.GetObjectKind().GroupVersionKind().GroupVersion().String(),
				},
			},
		}
		if err := k8sClient.Status().Update(ctx, kz); err != nil {
			t.Fatal(err)
		}

		_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(kz)})
		if err != nil {
			t.Fatal(err)
		}

		updated := &sourcev1alpha1.KustomizationSet{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kz), updated); err != nil {
			t.Fatal(err)
		}
	})
}

func deleteAllKustomizations(t *testing.T, cl client.Client) {
	t.Helper()
	err := cl.DeleteAllOf(context.TODO(), &kustomizev1.Kustomization{}, client.InNamespace("default"))
	if client.IgnoreNotFound(err) != nil {
		t.Fatal(err)
	}
}

func assertResourceDoesNotExist(t *testing.T, cl client.Client, ks *kustomizev1.Kustomization) {
	t.Helper()
	check := &kustomizev1.Kustomization{}
	if err := cl.Get(context.TODO(), client.ObjectKeyFromObject(ks), check); !apierrors.IsNotFound(err) {
		t.Fatalf("object %v still exists", ks)
	}
}

func assertKustomizationsExist(t *testing.T, cl client.Client, ns string, want ...string) {
	t.Helper()
	kss := &kustomizev1.KustomizationList{}
	if err := cl.List(context.TODO(), kss, client.InNamespace(ns)); err != nil {
		t.Fatalf("failed to list kustomizations in ns %s: %s", ns, err)
	}
	existingNames := func(l []kustomizev1.Kustomization) []string {
		names := []string{}
		for _, v := range l {
			names = append(names, v.GetName())
		}
		sort.Strings(names)
		return names
	}(kss.Items)

	sort.Strings(want)
	if diff := cmp.Diff(want, existingNames); diff != "" {
		t.Fatalf("got different names:\n%s", diff)
	}
}

func assertKustomizationCondition(t *testing.T, ks *sourcev1alpha1.KustomizationSet, condType, msg string) {
	cond := apimeta.FindStatusCondition(ks.Status.Conditions, condType)
	if cond == nil {
		t.Fatalf("failed to find matching status condition for type %s in %#v", condType, ks.Status.Conditions)
	}
	if cond.Message != msg {
		t.Fatalf("got %s, want %s", cond.Message, msg)
	}
}

func assertInventoryHasItems(t *testing.T, ks *sourcev1alpha1.KustomizationSet, objs ...runtime.Object) {
	t.Helper()
	if l := len(ks.Status.Inventory.Entries); l != len(objs) {
		t.Fatalf("expected %d items, got %v", len(objs), l)
	}
	entries := []sourcev1alpha1.ResourceRef{}
	for _, obj := range objs {
		objMeta, err := object.RuntimeToObjMeta(obj)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, sourcev1alpha1.ResourceRef{
			ID:      objMeta.String(),
			Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	want := &sourcev1alpha1.ResourceInventory{Entries: entries}
	if diff := cmp.Diff(want, ks.Status.Inventory); diff != "" {
		t.Fatalf("failed to get inventory:\n%s", diff)
	}
}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

func newKustomization(name, namespace string) *kustomizev1.Kustomization {
	return &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: 5 * time.Minute},
			Path:     "./examples/kustomize/environments/dev",
			Prune:    true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: "demo-repo",
			},
		},
	}
}

func newKustomizationSet(opts ...func(*sourcev1alpha1.KustomizationSet)) *sourcev1alpha1.KustomizationSet {
	ks := &sourcev1alpha1.KustomizationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-set",
			Namespace: "default",
		},
		Spec: sourcev1alpha1.KustomizationSetSpec{
			Generators: []sourcev1alpha1.KustomizationSetGenerator{
				{
					List: &sourcev1alpha1.ListGenerator{
						Elements: []apiextensionsv1.JSON{
							{Raw: []byte(`{"cluster": "engineering-dev"}`)},
							{Raw: []byte(`{"cluster": "engineering-prod"}`)},
							{Raw: []byte(`{"cluster": "engineering-preprod"}`)},
						},
					},
				},
			},
			Template: sourcev1alpha1.KustomizationSetTemplate{
				KustomizationSetTemplateMeta: sourcev1alpha1.KustomizationSetTemplateMeta{
					Name:      `{{.cluster}}-demo`,
					Namespace: "default",
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{Duration: 5 * time.Minute},
					Path:     "./clusters/{{.cluster}}/",
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: "GitRepository",
						Name: "demo-repo",
					},
					KubeConfig: &kustomizev1.KubeConfig{
						SecretRef: meta.SecretKeyReference{
							Name: "{{.cluster}}",
						},
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ks)
	}
	return ks
}
