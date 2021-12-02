package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/testenv"
	sourcev1alpha1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
)

const (
	timeout  = 10 * time.Second
	interval = 1 * time.Second
)

var (
	testEnv *testenv.Environment
	ctx     = ctrl.SetupSignalHandler()
)

func TestMain(m *testing.M) {
	utilruntime.Must(kustomizev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(sourcev1alpha1.AddToScheme(scheme.Scheme))
	testEnv = testenv.New(testenv.WithCRDPath(filepath.Join("..", "config", "crd", "bases")))

	if err := (&KustomizationSetReconciler{
		Client: testEnv,
	}).SetupWithManager(testEnv); err != nil {
		panic(fmt.Sprintf("Failed to start KustomizationSetReconciler: %v", err))
	}

	go func() {
		fmt.Println("Starting the test environment")
		if err := testEnv.Start(ctx); err != nil {
			panic(fmt.Sprintf("Failed to start the test environment manager: %v", err))
		}
	}()
	<-testEnv.Manager.Elected()

	code := m.Run()

	fmt.Println("Stopping the test environment")
	if err := testEnv.Stop(); err != nil {
		panic(fmt.Sprintf("Failed to stop the test environment: %v", err))
	}
	os.Exit(code)
}

func TestKustomizationSetReconciler_Reconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()

	obj := &sourcev1alpha1.KustomizationSet{
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
					Name: `{{cluster}}-demo`,
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{5 * time.Minute},
					Path:     "./clusters/{{cluster}}/",
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: "GitRepository",
						Name: "demo-repo",
					},
					KubeConfig: &kustomizev1.KubeConfig{
						SecretRef: meta.LocalObjectReference{
							Name: "{{cluster}}",
						},
					},
				},
			},
		},
	}
	g.Expect(testEnv.Create(ctx, obj)).To(Succeed())
	key := client.ObjectKey{Name: obj.Name, Namespace: obj.Namespace}

	var kustomizationList kustomizev1.KustomizationList
	g.Eventually(func() int {
		if err := testEnv.List(ctx, &kustomizationList, client.InNamespace(obj.Namespace)); err != nil {
			return 0
		}
		return len(kustomizationList.Items)
	}, timeout).Should(Equal(3))

	g.Expect(testEnv.Delete(ctx, obj)).To(Succeed())

	// Wait for GitRepository to be deleted
	g.Eventually(func() bool {
		if err := testEnv.Get(ctx, key, obj); err != nil {
			return apierrors.IsNotFound(err)
		}
		return false
	}, timeout).Should(BeTrue())
}
