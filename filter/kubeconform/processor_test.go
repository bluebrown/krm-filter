package kubeconform

import (
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
)

func TestFilter(t *testing.T) {
	checker := frameworktestutil.CommandResultsChecker{
		ErrorAssertionFunc: frameworktestutil.RequireStrippedStringsEqual,
		Command: func() *cobra.Command {
			return command.Build(Processor(), command.StandaloneEnabled, false)
		},
	}
	checker.Assert(t)
}
