# KRM Filter - KubeAudit

<!--mdtogo:Short-->

Audit kubernetes manifests

<!--mdtogo-->

<!--mdtogo:Long-->

The function uses the kubeadit module to validate and fix local kubernetes manifests.
It runs in either lint or fix mode.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

The function config is a simple configmap-like object where each data key-value
pair is one annotation.

```yaml
apiVersion: bluebrown.github.io/v1alpha1
kind: KubeAudit
metadata:
  name: security
# https://github.com/Shopify/kubeaudit#configuration-file
spec:
  # by default all auditors are enabled
  enabledAuditors: {}
  # specific auditor configs
  auditors:
    limits:
      cpu: "750m"
      memory: "500m"
```

Run the function as standalone providing the function config and resources.

```bash
kubeaudit fn-config.yaml - < resources.yaml
kubeaudit-fix fn-config.yaml - < resources.yaml
```

<!--mdtogo-->
