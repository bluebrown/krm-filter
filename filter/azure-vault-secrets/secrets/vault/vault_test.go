package vault

import (
	"os"
	"testing"
)

func TestSecretRetrievalSP(t *testing.T) {
	if os.Getenv("GO_INTEGRATION_TEST") != "1" {
		t.SkipNow()
	}

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
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
		{
			name:              "latest",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "",
			wantValue:         "latest",
		},
		{
			name:              "cached",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
	}

	cnfDir := "./testdata/sp"
	if err := os.MkdirAll(cnfDir, 0755); err != nil {
		t.Fatal(err)
	}

	sp, err := NewServicePrincipalFromEnv(cnfDir)
	if err != nil {
		t.Fatal(err)
	}

	src := NewSourceOrDie(cnfDir, sp, nil)

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
				t.Errorf("secret value does not match: want: %s got: %s", tt.wantValue, s)
			}
		})
	}

	err = os.RemoveAll(cnfDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecretRetrievalEnvToken(t *testing.T) {
	if os.Getenv("GO_INTEGRATION_TEST") != "1" {
		return
	}

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
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
		{
			name:              "latest",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "",
			wantValue:         "latest",
		},
		{
			name:              "cached",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
	}

	cnfDir := "./.testdata/env"
	if err := os.MkdirAll(cnfDir, 0755); err != nil {
		t.Fatal(err)
	}

	src := NewSourceOrDie(cnfDir, NewEnvTokenReader(), nil)

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
				t.Errorf("secret value does not match: want: %s got: %s", tt.wantValue, s)
			}
		})
	}

	err := os.RemoveAll(cnfDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecretRetrievalSPMemoryEnc(t *testing.T) {
	if os.Getenv("GO_INTEGRATION_TEST") != "1" {
		return
	}

	t.Parallel()

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
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
		{
			name:              "latest",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "",
			wantValue:         "latest",
		},
		{
			name:              "cached",
			giveVaultUri:      "https://krmtest.vault.azure.net/",
			giveSecretName:    "kustomize-plugin-testing",
			giveSecretVersion: "fb78a54ed06546a9ace9b261e95c795f",
			wantValue:         "version",
		},
	}

	cnfDir := "./.testdata/spmem"
	if err := os.MkdirAll(cnfDir, 0755); err != nil {
		t.Fatal(err)
	}

	enc := new(Encryption)

	sp, err := NewServicePrincipalFromEnv(cnfDir, WithEncryption(enc))
	if err != nil {
		t.Fatal(err)
	}

	src := NewSourceOrDie(cnfDir, sp, enc)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := src.RetrieveSecret(tt.giveVaultUri, tt.giveSecretName, tt.giveSecretVersion)
			if err != nil {
				t.Error(s)
				return
			}
			if s != tt.wantValue {
				t.Errorf("secret value does not match: want: %s got: %s", tt.wantValue, s)
			}
		})
	}

	err = os.RemoveAll(cnfDir)
	if err != nil {
		t.Fatal(err)
	}
}
