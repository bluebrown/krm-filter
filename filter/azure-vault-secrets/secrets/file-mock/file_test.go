package filemock

import (
	"testing"
)

func TestSecretRetrieval(t *testing.T) {
	tests := []struct {
		name              string
		giveVaultUri      string
		giveSecretName    string
		giveSecretVersion string
		wantValue         string
	}{
		{
			name:              "version",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "foo",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "bar",
		},
		{
			name:              "latest",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "foo",
			giveSecretVersion: "",
			wantValue:         "bar",
		},
		{
			name:              "envtoYaml",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "my-env-file",
			giveSecretVersion: "699e3794ed8f4ac799c08552d3d7a654",
			wantValue:         "FOO=bar\nBAZ=buz\nMORE=yes",
		},
	}

	src := NewSource("testdata")

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := src.RetrieveSecret(tt.giveVaultUri, tt.giveSecretName, tt.giveSecretVersion)
			if err != nil {
				t.Error(s)
				return
			}
			if s != tt.wantValue {
				t.Errorf("secret value does not match:\n---\nwant:\n%s\n---\ngot:\n%s", tt.wantValue, s)
			}
		})
	}

}
