package kubeaudit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/Shopify/kubeaudit"
	"github.com/Shopify/kubeaudit/auditors/all"
	"github.com/Shopify/kubeaudit/config"
	"github.com/bluebrown/krm-filter/util"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

type AuditFn func(audit *kubeaudit.Kubeaudit, path string, item *yaml.RNode, log util.LogFunc) error

// the audit func is used to run the processor in different modes. When passing
// in the validate func the processor acts as a validator. When passing in the
// transform func, the processor acts as a transformer fixing issues, if
// possible. After using the transformer, it is still recommend to use the
// validator since the transform cannot fix all possible errors
func Processor(fn AuditFn) framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		var (
			conf config.KubeauditConfig
			err  error
		)

		if rl.FunctionConfig != nil {
			if spec := rl.FunctionConfig.Field("spec"); spec != nil {
				conf, err = config.New(strings.NewReader(spec.Value.MustString()))
				if err != nil {
					return err
				}
			}
		}

		auditors, err := all.Auditors(conf)
		if err != nil {
			return fmt.Errorf("create auditors: %w", err)
		}

		audit, err := kubeaudit.New(auditors)
		if err != nil {
			return fmt.Errorf("init kubeaudit: %w", err)
		}

		return rl.Filter(kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
			errBuf := new(bytes.Buffer)

			for _, item := range items {
				logf := util.MakeLogFunc(rl, item)
				if err := fn(audit, util.MustGetPath(item), item, logf); err != nil {
					errBuf.WriteString(err.Error())
				}
			}

			errMsg := errBuf.String()
			return items, util.Ternary(len(errMsg) > 0,
				errors.New("\nThe following issues require manual intervention. Please fix these and run the function again:\n"+errMsg), nil)
		}))
	})
}

func Transform(audit *kubeaudit.Kubeaudit, path string, item *yaml.RNode, log util.LogFunc) error {
	report, err := audit.AuditManifest(path, strings.NewReader(item.MustString()))
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}

	if report == nil {
		return nil
	}

	results := report.Results()
	if len(results) == 0 {
		return nil
	}

	fixed := 0
	for _, result := range results {
		for _, auditResult := range result.GetAuditResults() {
			ok, plan := auditResult.FixPlan()
			if ok {
				log(framework.Info, "fixed: msg=%q tag=%q%s", plan, auditResult.Auditor, mapToKv(auditResult.Metadata))
				fixed++
			}
		}
	}

	if fixed == 0 {
		return nil
	}

	buf := new(bytes.Buffer)
	if err := report.Fix(buf); err != nil {
		return fmt.Errorf("fix: %w", err)
	}

	b, err := k8syaml.YAMLToJSON(buf.Bytes())
	if err != nil {
		return fmt.Errorf("yaml to json: %w", err)
	}

	if err := item.UnmarshalJSON(b); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}

func Validate(audit *kubeaudit.Kubeaudit, path string, item *yaml.RNode, log util.LogFunc) error {
	report, err := audit.AuditManifest(path, strings.NewReader(item.MustString()))
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	if report == nil {
		return nil
	}
	if len(report.Results()) == 0 {
		return nil
	}

	for _, res := range report.Results() {
		for _, ar := range res.GetAuditResults() {
			log(framework.Severity(ar.Severity.String()), "failed: msg=%q tag=%q%s", ar.Message, ar.Auditor, mapToKv(ar.Metadata))
		}
	}

	errBuf := new(bytes.Buffer)
	report.PrintResults(kubeaudit.WithColor(false), kubeaudit.WithWriter(errBuf))
	return errors.New(errBuf.String())
}

func mapToKv(m map[string]string) string {
	var buf bytes.Buffer
	for k, v := range m {
		if k == "MissingAnnotation" {
			continue
		}
		buf.WriteString(fmt.Sprintf(" %s=\"%s\"", strings.ToLower(k), v))
	}
	return buf.String()
}
