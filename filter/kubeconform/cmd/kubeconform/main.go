package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/kubeconform"
	"github.com/bluebrown/krm-filter/filter/kubeconform/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(kubeconform.Processor(), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.KubeconformShort
	cmd.Long = generated.KubeconformLong
	cmd.Example = generated.KubeconformExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
