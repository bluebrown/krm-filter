package azurevaultsecrets

import (
	"fmt"

	"github.com/bluebrown/krm-filter/util"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// returns a list of path matcher that match secret refs path relative to a pod
// this function should be called each time the matchers are needed as the
// PathMatcher store internal state that would carry over to the next run
func makeMatchers() []yaml.Filter {
	return []yaml.Filter{
		&yaml.PathMatcher{Path: []string{"spec", "containers", "*", "envFrom", "*", "secretRef", "name"}},
		&yaml.PathMatcher{Path: []string{"spec", "containers", "*", "env", "*", "valueFrom", "secretKeyRef", "name"}},
		&yaml.PathMatcher{Path: []string{"spec", "volumes", "*", "secret", "secretName"}},
	}
}

type HasSecretRef struct {
	Kind string
	// the secret name is the value to determine
	// if a resource has a reference to a secret
	SecretName string
}

func (f *HasSecretRef) Filter(node *yaml.RNode) (*yaml.RNode, error) {
	matchValue := yaml.MatchElement("", f.SecretName)

	pod, err := util.LookupPod(node)
	if err != nil {
		return node, err
	}

	if pod == nil {
		return nil, nil
	}

	for _, matchPath := range makeMatchers() {
		v, err := pod.Pipe(matchPath, matchValue)
		if err != nil {
			return node, err
		}
		if v != nil {
			return node, nil
		}
	}

	return nil, nil

}

type SecretRefInjector struct {
	Kind             string
	SecretName       string
	ContainerTargets []string
}

func (f *SecretRefInjector) Filter(node *yaml.RNode) (*yaml.RNode, error) {
	err := node.PipeE(
		yaml.LookupFirstMatch(yaml.ConventionalContainerPaths),
		matchSequence(nameMatcher(f.ContainerTargets...)),
		util.ForEach(func(node *yaml.RNode) error {
			return node.PipeE(
				yaml.LookupCreate(yaml.SequenceNode, "envFrom"),
				util.ReverseTee(
					&yaml.PathMatcher{Path: []string{"*", "secretRef", "name"}},
					yaml.MatchElement("", f.SecretName),
				),
				// TODO: find a better way to set the node
				yaml.Append(yaml.MustParse(fmt.Sprintf(`secretRef: { name: "%s" }`, f.SecretName)).YNode()),
			)
		}),
	)
	return node, err
}

// return the sequence if all list elements match the matcherFunc,
// otherwise return nil
func matchSequence(matcher framework.ResourceMatcherFunc) yaml.FilterFunc {
	return func(seq *yaml.RNode) (*yaml.RNode, error) {
		els, err := seq.Elements()
		if err != nil {
			return nil, err
		}

		rawSeq := &yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: make([]*yaml.Node, 0, len(els)),
		}

		for _, el := range els {
			if !matcher.Match(el) {
				continue
			}
			rawSeq.Content = append(rawSeq.Content, el.YNode())
		}

		return util.Ternary(len(rawSeq.Content) > 0, yaml.NewRNode(rawSeq), nil), nil
	}
}

func nameMatcher(names ...string) framework.ResourceMatcherFunc {
	return func(node *yaml.RNode) bool {
		n := node.Field("name")
		if n == nil {
			return false
		}
		return slices.Contains(names, yaml.GetValue(n.Value))
	}
}
