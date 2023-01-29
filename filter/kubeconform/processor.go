package kubeconform

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bluebrown/krm-filter/util"
	"github.com/yannh/kubeconform/pkg/resource"
	"github.com/yannh/kubeconform/pkg/validator"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const defaultSkips = "Kptfile,Kustomization,KubeConformValidator"

type Data struct {
	SchemaLocations      string `json:"schema_locations,omitempty"`
	schemaLocations      []string
	Cache                string `json:"cache,omitempty"`
	Debug                string `json:"debug,omitempty"`
	SkipTLS              string `json:"skip_tls,omitempty"`
	SkipKinds            string `json:"skip_kinds,omitempty"`
	RejectKinds          string `json:"reject_kinds,omitempty"`
	KubernetesVersion    string `json:"kubernetes_version,omitempty"`
	Strict               string `json:"strict,omitempty"`
	IgnoreMissingSchemas string `json:"ignore_missing_schemas,omitempty"`
}
type FunctionConfig struct {
	Data Data `yaml:"data,omitempty" json:"data,omitempty"`
}

func (v *FunctionConfig) AsOpts() validator.Opts {
	return validator.Opts{
		Cache:                v.Data.Cache,
		Debug:                v.Data.Debug == "true",
		SkipTLS:              v.Data.SkipTLS == "true",
		SkipKinds:            commaToMap(v.Data.SkipKinds),
		RejectKinds:          commaToMap(v.Data.RejectKinds),
		KubernetesVersion:    v.Data.KubernetesVersion,
		Strict:               v.Data.Strict == "true",
		IgnoreMissingSchemas: v.Data.IgnoreMissingSchemas == "true",
	}
}

func (v *FunctionConfig) Default() error {
	if v.Data.SkipKinds == "" {
		v.Data.SkipKinds = defaultSkips
	} else {
		v.Data.SkipKinds += "," + defaultSkips
	}
	if v.Data.KubernetesVersion == "" {
		v.Data.KubernetesVersion = "master"
	}
	if v.Data.Strict == "" {
		v.Data.Strict = "true"
	}
	versufffix := "{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{.ResourceKind}}{{.KindSuffix}}.json"
	v.Data.schemaLocations = []string{
		os.Getenv("KO_DATA_PATH") + "/schemas/" + versufffix,
		"https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/" + versufffix,
	}
	v.Data.schemaLocations = append(v.Data.schemaLocations, trimSplit(v.Data.SchemaLocations)...)
	return nil
}

func Processor() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		config := &FunctionConfig{}
		if err := framework.LoadFunctionConfig(rl.FunctionConfig, config); err != nil {
			return fmt.Errorf("read function config: %w", err)
		}

		v, err := validator.New(config.Data.schemaLocations, config.AsOpts())
		if err != nil {
			return fmt.Errorf("create validator: %w", err)
		}

		filter := KubeConform{
			validator: v,
		}

		_, err = filter.Each(rl.Items)
		rl.Results = append(rl.Results, filter.Results...)

		return err
	})
}

type KubeConform struct {
	validator validator.Validator
	Results   framework.Results
}

func (f *KubeConform) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	result := f.validator.ValidateResource(resource.Resource{
		Path:  object.GetAnnotations()[kioutil.PathAnnotation],
		Bytes: []byte(object.MustString()),
	})

	if result.Err == nil {
		return object, nil
	}

	var errs error
	for _, err := range result.ValidationErrors {
		msg := fmt.Sprintf("[%s] %s: %s", statusToString(result.Status), err.Path, err.Msg)
		errs = errors.Join(errs, fmt.Errorf("%s: %s", util.ResourceID(object), msg))
		f.Results = append(f.Results, util.ItemToResult(
			object,
			util.Ternary(result.Status == validator.Error, framework.Error, framework.Warning),
			msg,
		))
	}

	return nil, errs
}

func (f *KubeConform) Each(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var err error
	for _, item := range items {
		err = errors.Join(err, item.PipeE(f))
	}
	return items, err
}

func commaToMap(s string) map[string]struct{} {
	items := strings.Split(s, ",")
	m := make(map[string]struct{}, len(items))
	for _, item := range items {
		m[strings.TrimSpace(item)] = struct{}{}
	}
	return m
}

func trimSplit(s string) []string {
	ss := strings.Split(s, ",")
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}
	return ss
}

func statusToString(status validator.Status) string {
	switch status {
	case validator.Error:
		return "error"
	case validator.Skipped:
		return "skipped"
	case validator.Valid:
		return "valid"
	case validator.Invalid:
		return "invalid"
	case validator.Empty:
		return "empty"
	default:
		return "unknown"
	}
}
