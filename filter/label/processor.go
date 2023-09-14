package label

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
		ta := &TopLevelLabeler{
			Labels: conf.Data,
		}
		return rl.Filter(kio.FilterFunc(ta.Each))
	})
}

type TopLevelLabeler struct {
	Labels map[string]string
}

func (a *TopLevelLabeler) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	labels := object.GetLabels()
	for k, v := range a.Labels {
		labels[k] = v
	}
	if err := object.SetLabels(labels); err != nil {
		return object, err
	}
	return object, nil
}

func (a *TopLevelLabeler) Each(r []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, object := range r {
		if err := object.PipeE(a); err != nil {
			return r, err
		}
	}
	return r, nil
}
