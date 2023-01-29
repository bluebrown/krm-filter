package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/annotate"
	"github.com/bluebrown/krm-filter/filter/annotate/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(annotate.Processor(), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.AnnotateShort
	cmd.Long = generated.AnnotateLong
	cmd.Example = generated.AnnotateExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
