# KRM Filter - Patchwork

<!--mdtogo:Short-->

Kustomize patching on steroids

<!--mdtogo-->

<!--mdtogo:Long-->

An alternative way to patch resources using kustomize's replacement selector.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

The function config is a custom resourcce of kind `Patchwork`.

```yaml
apiVersion: bluebrown.github.io/v1alpha1
kind: Patchwork
metadata:
  name: inject-sidecar
spec:
  patches:
    - resource:
        # pre filter based on kind
        # can be a regex pattern
        kind: (Deployment|Job|CronJob)
      # look up the first match of the paths
      # further selector queries are relative to it
      lookup:
        - spec.template
        - spec.jobTemplate.spec.template
      values:
        # apply one or more patches to the given
        # paths relative to lookup
        - paths:
            - spec.containers.[name=sidecar]
          value:
            name: sidecar
            image: busybox
            command:
              - sleep
            args:
              - infinity
```

Run the function as standalone providing the function config and resources.

```bash
patchwork fn-config.yaml - < resources.yaml
```

<!--mdtogo-->
