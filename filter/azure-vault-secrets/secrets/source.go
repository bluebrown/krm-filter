package secrets

import (
	"fmt"
	"os"
	"path"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	filemock "github.com/bluebrown/krm-filter/filter/azure-vault-secrets/secrets/file-mock"
	"github.com/bluebrown/krm-filter/filter/azure-vault-secrets/secrets/vault"
	"github.com/bluebrown/krm-filter/util"
)

type SourceKind string

const (
	SourceKindAzure    SourceKind = "azure"
	SourceKindFileMock SourceKind = "file-mock"
)

type Source interface {
	RetrieveSecret(uri, name, version string) (string, error)
}

type VaultSecret struct {
	Key     string `json:"key" yaml:"key"`
	Secret  string `json:"secret" yaml:"secret"`
	Version string `json:"version" yaml:"version"`
}

type SecretFetcher struct {
	Fs     filesys.FileSystemOrOnDisk
	Source Source
}

type SourceFlags uint8

const (
	DISK_FS SourceFlags = 1 << iota
	FILE_MOCK
)

func NewSecretFetcher(options SourceFlags) (*SecretFetcher, error) {
	fs := filesys.FileSystemOrOnDisk{}
	if options&DISK_FS < 1 {
		fs.FileSystem = filesys.MakeFsInMemory()
	}

	var src Source
	if options&FILE_MOCK > 0 {
		src = filemock.NewSource(os.Getenv("FILE_MOCK_DATA_DIR"))
	} else {
		enc := vault.NewEnc(fs)
		homeConf, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		configDir := path.Join(homeConf, "krm-filter", "azure-vault-secrets")
		var ts vault.TokenSource
		if t := os.Getenv("AAD_ACCESS_TOKEN"); t != "" {
			ts = vault.NewEnvTokenReader()
		} else {
			ts, err = vault.NewServicePrincipalFromEnv(configDir, vault.WithEncryption(enc))
			if err != nil {
				return nil, err
			}
		}
		src = vault.NewSourceOrDie(configDir, ts, enc)
	}

	return &SecretFetcher{
		Source: src,
		Fs:     fs,
	}, nil
}

func (f *SecretFetcher) FetchSecrets(vaultUri string, secretSpecs []VaultSecret) (map[string]string, error) {
	templateContext := make(map[string]string, len(secretSpecs))
	for _, s := range secretSpecs {
		v, err := f.Source.RetrieveSecret(vaultUri, s.Secret, s.Version)
		if err != nil {
			return nil, fmt.Errorf("fetch %s, %s, %s: %w",
				vaultUri, s.Secret, s.Version, err,
			)
		}
		templateContext[util.Ternary(s.Key != "", s.Key, s.Secret)] = v
	}
	return templateContext, nil
}
