package util

import "sigs.k8s.io/kustomize/kyaml/yaml"

// ReverseTee is like tee in that it passes the arguments instead the output,
// but it will only pass the arguments, if the filter pipe returns nil.
// this can be useful when searching for the right place to perform an operation
// and then cancel early without running subsequent pipes
func ReverseTee(filters ...yaml.Filter) yaml.Filter {
	return ReverseTeePiper{Filters: filters}
}

type ReverseTeePiper struct {
	Kind string `yaml:"kind,omitempty"`

	// Filters are the set of Filters run by ReverseTeePiper.
	Filters []yaml.Filter `yaml:"filters,omitempty"`
}

func (t ReverseTeePiper) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	v, err := rn.Pipe(t.Filters...)
	if err != nil {
		return rn, err
	}
	if v != nil {
		return nil, nil
	}
	return rn, nil
}
