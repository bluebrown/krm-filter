package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/label"
	"github.com/bluebrown/krm-filter/filter/label/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(label.Processor(), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.LabelShort
	cmd.Long = generated.LabelLong
	cmd.Example = generated.LabelExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
