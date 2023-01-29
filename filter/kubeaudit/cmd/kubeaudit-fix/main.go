package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"github.com/bluebrown/krm-filter/filter/kubeaudit"
	"github.com/bluebrown/krm-filter/filter/kubeaudit/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(kubeaudit.Processor(kubeaudit.Transform), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.KubeauditShort
	cmd.Long = generated.KubeauditLong
	cmd.Example = generated.KubeauditExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
