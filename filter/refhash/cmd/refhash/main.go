package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/refhash"
	"github.com/bluebrown/krm-filter/filter/refhash/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(refhash.Processor(), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.RefhashShort
	cmd.Long = generated.RefhashLong
	cmd.Example = generated.RefhashExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
