package annotate

import (
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type FunctionConfig struct {
	Data map[string]string
}

func Processor() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		conf := new(FunctionConfig)
		if err := framework.LoadFunctionConfig(rl.FunctionConfig, conf); err != nil {
			return err
		}
		ta := &TopLevelAnnotator{
			Annotations: conf.Data,
		}
		return rl.Filter(kio.FilterFunc(ta.Each))
	})
}

type TopLevelAnnotator struct {
	Annotations map[string]string
}

func (a *TopLevelAnnotator) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	annos := object.GetAnnotations()
	for k, v := range a.Annotations {
		annos[k] = v
	}
	if err := object.SetAnnotations(annos); err != nil {
		return object, err
	}
	return object, nil
}

func (a *TopLevelAnnotator) Each(r []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, object := range r {
		if err := object.PipeE(a); err != nil {
			return r, err
		}
	}
	return r, nil
}
