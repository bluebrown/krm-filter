# KRM Filter - KubeConform

<!--mdtogo:Short-->

kubernetes manifest schema validation

<!--mdtogo-->

<!--mdtogo:Long-->

Use kubeconform to perform schema validation against local kubernetes manifests

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
  cache: ""
  debug: false
  skip_tls: false
  skip_kinds: "" # comma separated list
  reject_kinds: "" # comma separated list
  kubernetes_version: "master"
  strict: false
  ignore_missing_schemas: false
```

Run the function as standalone providing the function config and resources.

```bash
kubeconform fn-config.yaml - < resources.yaml
```

<!--mdtogo-->
