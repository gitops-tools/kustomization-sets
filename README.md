# kustomize-set-controller

## Building

```shell
export IMG=<insert docker registry>
$ make docker-build docker-push
```

## Installation

This requires the source-controller and kustomize-controller.

```shell
$ flux install --components source-controller,kustomize-controller
$ IMG=<insert docker registry> make deploy
```

You'll need a Flux `GitRepository` [object](https://fluxcd.io/docs/components/source/gitrepositories/).

For example:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: go-demo-repo
  namespace: default
spec:
  interval: 15m
  url: https://github.com/bigkevmcd/go-demo
```

And then create a `KustomizationSet`.

For example:

```yaml
apiVersion: source.gitops.solutions/v1alpha1
kind: KustomizationSet
metadata:
  name: go-demo-set
  namespace: default
spec:
  generators:
  - list:
      elements:
      - env: dev
      - env: production
      - env: staging
  template:
    metadata:
      name: '{{env}}-demo'
      namespace: default
    spec:
      interval: 5m
      path: "./examples/kustomize/environments/{{ env }}"
      prune: true
      sourceRef:
        kind: GitRepository
        name: go-demo-repo
```

This will trigger the deployment of the three environments in the repo above.
