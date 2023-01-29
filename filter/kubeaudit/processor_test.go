package kubeaudit

import (
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
)

func TestFilterFix(t *testing.T) {
	checker := frameworktestutil.CommandResultsChecker{
		TestDataDirectory: "testdata/fix",
		Command: func() *cobra.Command {
			return command.Build(Processor(Transform), command.StandaloneEnabled, false)
		},
	}
	checker.Assert(t)
}

func TestFilterAudit(t *testing.T) {
	checker := frameworktestutil.ProcessorResultsChecker{
		TestDataDirectory:  "testdata/audit",
		ErrorAssertionFunc: frameworktestutil.RequireStrippedStringsEqual,
		Processor: func() framework.ResourceListProcessor {
			return Processor(Validate)
		},
	}
	checker.Assert(t)
}
