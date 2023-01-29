package patchwork

import (
	"encoding/json"
	"regexp"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/utils"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

type Resource struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
}

type Value struct {
	Paths []string
	Value json.RawMessage
}

type Patch struct {
	Resource Resource `json:"resource,omitempty"`
	Lookup   []string `json:"lookup,omitempty"`
	Values   []Value
}

type PatchSpec struct {
	Patches []Patch
}

type FunctionConfig struct {
	Spec PatchSpec `json:"spec,omitempty"`
}

type Processor struct{}

func (p *Processor) Process(rl *framework.ResourceList) error {
	conf := FunctionConfig{}
	if err := framework.LoadFunctionConfig(rl.FunctionConfig, &conf); err != nil {
		return err
	}

	for _, patch := range conf.Spec.Patches {

		// normalize the lookup paths
		// these path wont be created if not exist
		// if there are no lookup path and empty string
		// is used to match the root of the document
		var paths [][]string
		if len(patch.Lookup) == 0 {
			paths = [][]string{{""}}
		} else {
			paths = make([][]string, len(patch.Lookup))
			for i, p := range patch.Lookup {
				paths[i] = utils.SmarterPathSplitter(p, ".")
			}
		}

		// filter the items
		for _, item := range rl.Items {
			object, err := item.Pipe(
				MatchField("kind", patch.Resource.Kind),
				MatchField("metadata.name", patch.Resource.Name),
				yaml.LookupFirstMatch(paths),
			)
			if err != nil {
				return err
			}
			if object == nil {
				continue
			}

			for _, val := range patch.Values {
				b, err := k8syaml.JSONToYAML(val.Value)
				if err != nil {
					return err
				}

				node, err := yaml.Parse(string(b))
				if err != nil {
					return err
				}

				filters := []yaml.Filter{}
				for _, path := range val.Paths {
					filters = append(filters, yaml.Tee(
						&yaml.PathMatcher{
							Path:   utils.SmarterPathSplitter(path, "."),
							Create: node.YNode().Kind,
						},
						yaml.FilterFunc(func(object *yaml.RNode) (*yaml.RNode, error) {
							return object, object.VisitElements(func(n *yaml.RNode) error {
								n.SetYNode(node.Copy().YNode())
								return nil
							})
						}),
					))
				}

				if err := object.PipeE(filters...); err != nil {
					return err
				}

			}
		}

	}

	return nil

}

func MatchField(field string, pattern string) yaml.FilterFunc {
	return func(object *yaml.RNode) (*yaml.RNode, error) {
		if pattern == "" {
			return object, nil
		}
		f, err := object.Pipe(yaml.Lookup(utils.SmarterPathSplitter(field, ".")...))
		if err != nil {
			return nil, err
		}
		if f == nil {
			return nil, nil
		}
		ok, err := regexp.MatchString(pattern, f.YNode().Value)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
		return object, nil
	}
}

func CopyValue(val *yaml.RNode) yaml.FilterFunc {
	return func(object *yaml.RNode) (*yaml.RNode, error) {
		object.SetYNode(val.Copy().YNode())
		return object, nil
	}
}
