package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	azurevaultsecrets "github.com/bluebrown/krm-filter/filter/azure-vault-secrets"
	"github.com/bluebrown/krm-filter/filter/azure-vault-secrets/generated"
)

var (
	version  = "devel"
	revision = "unknown"
)

func main() {
	cmd := command.Build(azurevaultsecrets.Processor(), command.StandaloneEnabled, false)

	cmd.Version = fmt.Sprintf("version=%q revision=%q", version, revision)
	cmd.Short = generated.AzureVaultSecretsShort
	cmd.Long = generated.AzureVaultSecretsLong
	cmd.Example = generated.AzureVaultSecretsExamples

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
