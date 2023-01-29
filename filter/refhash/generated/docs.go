

// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var RefhashShort = `Find references to secrets and configmaps and annotate their holder with
checksums`
var RefhashLong = `
Find all references to resource and annotate their holder with a checksum of the
found reference. This allows to trigger a new deployment rollout or similar pod
restarts when the content of the references resource has changed-
`
var RefhashExamples = `
The function config is a simple configmap-like object containing options.

  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: hash
  data:
    secret_kinds: Secret,SealedSecret
    configmap_kinds: ConfigMap

Run the function as standalone providing the function config and resources.

  refhash fn-config.yaml - < resources.yaml
`