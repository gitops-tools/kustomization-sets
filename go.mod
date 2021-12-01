module github.com/gitops-tools/kustomize-set-controller

go 1.16

require (
	github.com/fluxcd/kustomize-controller/api v0.18.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)
