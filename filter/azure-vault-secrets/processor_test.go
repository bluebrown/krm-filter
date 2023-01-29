package azurevaultsecrets

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
)

func TestFilter(t *testing.T) {
	os.Setenv("FILE_MOCK_DATA_DIR", "file-mock")
	checker := frameworktestutil.CommandResultsChecker{
		TestDataDirectory: "testdata/cases/",
		Command: func() *cobra.Command {
			return command.Build(Processor(), command.StandaloneEnabled, false)
		},
	}
	checker.Assert(t)
}
