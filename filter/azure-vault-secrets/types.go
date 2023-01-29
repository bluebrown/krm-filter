package azurevaultsecrets

import "github.com/bluebrown/krm-filter/filter/azure-vault-secrets/secrets"

type ModeKind string

const (
	ModeKindAzure ModeKind = "azure"
	ModeKindMock  ModeKind = "file-mock"
)

type FsKind string

const (
	FsKindMemory FsKind = "memory"
	FsKindDisk   FsKind = "disk"
)

type FunctionConfig struct {
	Data Data `json:"data,omitempty"`
}

type Data struct {
	Mode ModeKind
	Fs   FsKind
}

type AzureVaultSourceSpec struct {
	VaultUri         string                `json:"vaultUri" yaml:"vaultUri"`
	VaultSecrets     []secrets.VaultSecret `json:"vaultSecrets" yaml:"vaultSecrets"`
	ContainerTargets []string              `json:"containerTargets" yaml:"containerTargets"`
	SecretName       string                `json:"secretName" yaml:"secretName"`
	Template         *string               `json:"stringDataTemplate" yaml:"stringDataTemplate"`
}

type AzureVaultSourceCRD struct {
	Spec AzureVaultSourceSpec `json:"spec,omitempty"`
}

type secretRef struct {
	name       string
	data       map[string]string
	template   *string
	containers []string
	nss        Namespaces
	annos      map[string]string
}
