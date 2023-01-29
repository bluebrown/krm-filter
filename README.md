# KRM Filter

This mono repo contains krm filter (functions) that conform to the [krm
functions
spec](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md).

These filter can be used as standalone or managed by tools that support krm
filter. Currently known tools are
[kustomize](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/containerized_krm_functions/)
and [kpt](https://kpt.dev/book/04-using-functions/).

## Example

```bash
kustomize fn run --image bluebrown/kubeaudit-fix dir/
```

## Standalone

The filter can be run in standalone mode since they are wrapped as [cobra
commands](https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml@v0.13.10/fn/framework/command#Build).

Keep in mind that you MUST provide a config as first positional argument for
each filter.

```bash
cat resources.yaml \
  | kubeconform schema-conf.yaml - \
  | kubeadit-fix audit-conf.yaml - \
  | kubeadit audit-conf.yaml - \
```

## Resource List

With the help of some tools, i.e. kpt, which can generate a resource list,
the filter can be chained together as well.

```bash
kpt fn source inputdir \
  | kubeconform \
  | kubeaudit-fix \
  | kubeaudit \
  | kpt fn sink outputdir
```

However, this way, there is no way of providing the function config. So it if that is required
it is better to go all in by using kpt fn eval.

```bash
kpt fn source inputdir \
  | kpt fn eval - -i ko.local/kubeconform --network -- strict=true \
  | kpt fn eval - -i ko.local/kubeaudit-fix --fn-config audit.yaml \
  | kpt fn eval - -i ko.local/kubeaudit --fn-config audit.yaml \
  | kpt fn sink outputdir
```

## Development

```bash
bin/init my-new-filter
```
