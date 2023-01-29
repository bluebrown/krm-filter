package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	KindPod = "Pod"
)

type LogFunc func(sev framework.Severity, fstr string, args ...any)

func MakeLogFunc(rl *framework.ResourceList, item *yaml.RNode) LogFunc {
	return func(sev framework.Severity, fstr string, args ...any) {
		rl.Results = append(rl.Results, ItemToResult(item, sev, fmt.Sprintf(fstr, args...)))
	}
}

func Ternary[T any](cond bool, yes T, no T) T {
	if cond {
		return yes
	} else {
		return no
	}
}

func GVK(object *yaml.RNode) string {
	return fmt.Sprintf("%s/%s/%s",
		object.GetApiVersion(), object.GetKind(), object.GetName(),
	)
}

func ResourceID(object *yaml.RNode) string {
	s := fmt.Sprintf("%s(%d):",
		object.GetAnnotations()[kioutil.PathAnnotation],
		mustAtoi(object.GetAnnotations()[kioutil.IndexAnnotation]),
	)
	if ns := object.GetNamespace(); ns != "" {
		s = fmt.Sprintf("%s [%s]:", s, ns)
	}

	s = fmt.Sprintf("%s %s", s, GVK(object))

	return s
}

func ItemToResult(item *yaml.RNode, sev framework.Severity, msg string) *framework.Result {
	return &framework.Result{
		Severity: sev,
		Message:  msg,
		File: &framework.File{
			Path:  item.GetAnnotations()[kioutil.PathAnnotation],
			Index: mustAtoi(item.GetAnnotations()[kioutil.IndexAnnotation]),
		},
		ResourceRef: &yaml.ResourceIdentifier{
			TypeMeta: yaml.TypeMeta{
				APIVersion: item.GetApiVersion(),
				Kind:       item.GetKind(),
			},
			NameMeta: yaml.NameMeta{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		},
	}
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func MustGetPath(item *yaml.RNode) string {
	return item.GetAnnotations()[kioutil.PathAnnotation]
}

// forEach returns a filter that calls visitElements with the "do"
// function on the input sequence node
func ForEach(do func(node *yaml.RNode) error) yaml.Filter {
	return yaml.FilterFunc(func(object *yaml.RNode) (*yaml.RNode, error) {
		return object, object.VisitElements(do)
	})
}

// CSV can be used to use a string slice as data while reading it from the
// function config data as comma separated string. This is useful when using
// simple configmaps as function config as those don't support anything other
// than a string map
type CSV []string

func (dst *CSV) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*dst = strings.Split(s, ",")
	return nil
}

// list of conventional pod template path
var podPath = [][]string{
	{"spec", "template"},
	{"spec", "jobTemplate", "spec", "template"},
	{"template"},
}

// get the pod from the input node. If the node is of kind pod,
// the node itself is returned. Otherwise, the pod template from
// a list of conventional path is looked up.
func LookupPod(object *yaml.RNode) (*yaml.RNode, error) {
	if object.GetKind() == KindPod {
		return object, nil
	}
	return yaml.LookupFirstMatch(podPath).Filter(object)
}
