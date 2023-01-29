# KRM Filter - Annotate

<!--mdtogo:Short-->

Add top-level annotations to all resources.

<!--mdtogo-->

<!--mdtogo:Long-->

Sometimes annotations contain only meta information and, for example, a pod
restart is not desired. This function annotates only the top level metadata of
each resource, in order to avoid pod restarts and problems with immutable
fields.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

The function config is a simple configmap-like object where each data key-value
pair is one annotation.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: annotate
data:
  repo: https://github.com/bluebrown/app
  revision: v1
```

Run the function as standalone providing the function config and resources.

```bash
annotate fn-config.yaml - < resources.yaml
```

<!--mdtogo-->
