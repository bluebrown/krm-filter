# KRM Filter - Annotate

<!--mdtogo:Short-->

Add top-level labels to all resources.

<!--mdtogo-->

<!--mdtogo:Long-->

Sometimes labels contain only meta information and, for example, a pod
restart is not desired. This function labels only the top level metadata of
each resource, in order to avoid pod restarts and problems with immutable
fields.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

The function config is a simple configmap-like object where each data key-value
pair is one label.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: label
data:
  app.kubernetes.io/name: nginx
  app.kubernetes.io/component: ingress-controller
```

Run the function as standalone providing the function config and resources.

```bash
label fn-config.yaml - < resources.yaml
```

<!--mdtogo-->
