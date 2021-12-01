module github.com/gitops-tools/kustomize-set-controller

go 1.16

require (
	github.com/fluxcd/kustomize-controller/api v0.18.1
	github.com/fluxcd/pkg/apis/meta v0.10.1
	github.com/imdario/mergo v0.3.12
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/stretchr/testify v1.7.0
	github.com/valyala/fasttemplate v1.2.1
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)
