module github.com/gitops-tools/kustomize-set-controller

go 1.16

require (
	github.com/fluxcd/kustomize-controller/api v0.18.1
	github.com/fluxcd/pkg/apis/meta v0.10.2
	github.com/fluxcd/pkg/runtime v0.12.2
	github.com/fluxcd/source-controller/api v0.21.1
	github.com/go-logr/logr v1.2.2
	github.com/google/go-cmp v0.5.6
	github.com/imdario/mergo v0.3.12
	github.com/jenkins-x/go-scm v1.10.11
	github.com/onsi/gomega v1.17.0
	github.com/valyala/fasttemplate v1.2.1
	k8s.io/api v0.23.0
	k8s.io/apiextensions-apiserver v0.23.0
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.0
	sigs.k8s.io/controller-runtime v0.11.0
)
