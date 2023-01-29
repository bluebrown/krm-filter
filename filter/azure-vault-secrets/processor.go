package azurevaultsecrets

import (
	"bytes"
	"fmt"

	"github.com/bluebrown/krm-filter/filter/azure-vault-secrets/secrets"
	"github.com/bluebrown/krm-filter/util"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	KindPod         = "Pod"
	KindSecret      = "Secret"
	KindVaultSource = "AzureVaultSource"
)

func Processor() framework.ResourceListProcessorFunc {
	return func(rl *framework.ResourceList) error {
		config := &FunctionConfig{}
		if err := framework.LoadFunctionConfig(rl.FunctionConfig, config); err != nil {
			return err
		}

		var flags secrets.SourceFlags

		if config.Data.Mode == ModeKindMock {
			flags |= secrets.FILE_MOCK
		}

		if config.Data.Fs == FsKindDisk {
			flags |= secrets.DISK_FS
		}

		sf, err := secrets.NewSecretFetcher(flags)
		if err != nil {
			return err
		}

		// first find all the custom resources to generate secrets from
		crs, err := framework.KindMatcher(KindVaultSource).Filter(rl.Items)
		if err != nil {
			return err
		}

		// refs is used to store the secret data retrieved from the vault
		refs := make([]secretRef, len(crs))

		// fetch the secrets and capture the reference meta data
		// such as the name and container targets
		for i, cr := range crs {
			avs := AzureVaultSourceCRD{}
			if err := cr.YNode().Decode(&avs); err != nil {
				return err
			}

			data, err := sf.FetchSecrets(avs.Spec.VaultUri, avs.Spec.VaultSecrets)
			if err != nil {
				return err
			}

			annos := make(map[string]string, len(avs.Spec.VaultSecrets)+1)
			annos["keyvault/uri"] = avs.Spec.VaultUri

			for _, s := range avs.Spec.VaultSecrets {
				annos["version/"+s.Secret] = util.Ternary(s.Version != "", s.Version, "latest")
			}

			refs[i] = secretRef{
				name:       avs.Spec.SecretName,
				data:       data,
				template:   avs.Spec.Template,
				containers: avs.Spec.ContainerTargets,
				nss:        make(Namespaces),
				annos:      annos,
			}

		}

		// inject a ref into the container targets where applicable.
		// And find all references to all vault secrets the entire resource list.
		//
		// TODO: this could be more efficient, altering the pipeline algo
		// for example once secrets refs in a container are found, they could be compared to all refs
		// instead of searching each time from the top of the resource
		for _, item := range rl.Items {
			for _, r := range refs {

				injectSecretRef := &SecretRefInjector{SecretName: r.name, ContainerTargets: r.containers}
				hasSecretRef := &HasSecretRef{SecretName: r.name}

				notInNss := yaml.FilterFunc(r.nss.NotFound)
				addToNss := yaml.FilterFunc(r.nss.Add)

				err := item.PipeE(
					// first try to inject the secret and remember the namespace,
					// if the ref was injected
					util.ReverseTee(injectSecretRef, addToNss),
					// otherwise, check if the secret is in the given namespace,
					// and if has already a secret ref. If so, add it to the namespaces
					util.ReverseTee(notInNss, hasSecretRef, addToNss),
				)

				if err != nil {
					return err
				}

			}
		}

		// finally generate all the secrets, once the target namespaces are known.
		// the namespaces are only know now because they are derived from the resources
		// containing either containerTargets or reference the secret by itself
		for _, r := range refs {
			gen := SecretGenerator{
				SecretName:  r.name,
				StringData:  r.data,
				Namespaces:  r.nss,
				Template:    r.template,
				Annotations: r.annos,
			}

			rl.Items, err = gen.Filter(rl.Items)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

type Namespaces map[string]struct{}

func (nss Namespaces) NotFound(object *yaml.RNode) (*yaml.RNode, error) {
	if _, ok := nss[object.GetNamespace()]; ok {
		return nil, nil
	}
	return object, nil
}

func (nss Namespaces) Add(object *yaml.RNode) (*yaml.RNode, error) {
	nss[object.GetNamespace()] = struct{}{}
	return object, nil
}

const secretStr = `---
apiVersion: v1
kind: Secret
metadata:
    name: %s
type: Opaque
data: {}
`

func createSecret(name string, stringData map[string]string) (*yaml.RNode, error) {
	yn, err := yaml.Parse(fmt.Sprintf(secretStr, name))
	if err != nil {
		return nil, err
	}
	err = yn.LoadMapIntoSecretData(stringData)
	return yn, err
}

type SecretGenerator struct {
	Kind string `yaml:"kind,omitempty" json:"kind,omitempty"`

	Namespaces  Namespaces        `json:"namespaces,omitempty"`
	SecretName  string            `json:"secret_name,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	StringData  map[string]string `json:"string_data,omitempty"`
	Template    *string           `json:"template,omitempty"`
}

func (g SecretGenerator) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var data map[string]string

	if g.Template != nil {
		data = make(map[string]string)

		t, err := TemplateRenderer.Parse(*g.Template)
		if err != nil {
			return items, err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, g.StringData); err != nil {
			return items, err
		}
		if err := yaml.Unmarshal(buf.Bytes(), &data); err != nil {
			return items, err
		}
	} else {
		data = g.StringData
	}

	secret, err := createSecret(g.SecretName, data)
	if err != nil {
		return items, err
	}

	if err := secret.SetAnnotations(g.Annotations); err != nil {
		return items, err
	}

	if err := secret.SetLabels(g.Labels); err != nil {
		return items, err
	}

	// filter out existing secrets in order to avoid generating duplicates
	newItems := make([]*yaml.RNode, 0, len(items))
	for _, item := range items {
		if item.GetName() != g.SecretName {
			newItems = append(newItems, item)
			continue
		}
		if item.GetKind() != KindSecret {
			newItems = append(newItems, item)
			continue
		}
		if _, ok := g.Namespaces[item.GetNamespace()]; !ok {
			newItems = append(newItems, item)
			continue
		}
	}

	for ns := range g.Namespaces {
		s := secret.Copy()
		if err := s.SetNamespace(ns); err != nil {
			return items, err
		}
		items = append(newItems, s)
	}

	return items, nil
}
