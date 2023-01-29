package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/patchwork"
	"github.com/bluebrown/krm-filter/filter/patchwork/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(&patchwork.Processor{}, command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.PatchworkShort
	cmd.Long = generated.PatchworkLong
	cmd.Example = generated.PatchworkExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
