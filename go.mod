module github.com/gitops-tools/kustomize-set-controller

go 1.16

require (
	github.com/fluxcd/kustomize-controller/api v0.18.1
	github.com/fluxcd/pkg/apis/meta v0.10.1
	github.com/fluxcd/pkg/runtime v0.12.2
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/google/go-cmp v0.5.5
	github.com/imdario/mergo v0.3.12
	github.com/jenkins-x/go-scm v1.10.11
	github.com/onsi/gomega v1.15.0
	github.com/valyala/fasttemplate v1.2.1
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)
