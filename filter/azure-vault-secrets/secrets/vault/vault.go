package vault

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

// TODO: use proper context

type TokenSource interface {
	AccessToken() (string, error)
	Salt() (string, error)
}

type Source struct {
	configDir string
	tok       TokenSource
	salt      string
	enc       *Encryption
}

func NewSourceOrDie(cnfDir string, tokenSrc TokenSource, enc *Encryption) Source {
	s, err := tokenSrc.Salt()
	if err != nil {
		panic(err)
	}

	if enc == nil {
		enc = new(Encryption)
	}

	return Source{configDir: cnfDir, tok: tokenSrc, salt: s, enc: enc}
}

func (src Source) RetrieveSecret(uri, name, version string) (string, error) {

	var (
		secretHash string
		secretPath string
	)

	// check first if the secret is in cache and use it from there
	// only secret with a specific version are cached

	if version != "" {
		secretHash = getHash(uri, name, version, src.salt)
		secretPath = path.Join(src.configDir, secretHash)
		cachedSecret, err := src.enc.DecryptAtRest(secretPath, src.salt)
		if err == nil {
			// return the secret if it exists in cache
			return string(cachedSecret), nil
		} else if !errors.Is(err, os.ErrNotExist) {
			// otherwise return with error for any other error than IsnotExist
			return "", fmt.Errorf("error while reading cached secret: %w", err)
		}
	}

	// finally, fetch the secret from vault.
	token, err := src.tok.AccessToken()
	if err != nil {
		return "", fmt.Errorf("could not get token: %w", err)
	}

	secretRes, err := getSecretFromVault(context.TODO(), token, uri, name, version)
	if err != nil {
		return "", fmt.Errorf("could not fetch secret: %w", err)
	}

	// TODO: use pointer to check against nil
	if version == "" {
		return secretRes.Value, nil
	}
	err = src.enc.EncryptAtRest(secretPath, []byte(secretRes.Value), src.salt)
	if err != nil {
		return "", fmt.Errorf("could not cache secret: %w", err)
	}

	return secretRes.Value, nil
}

func getHash(ss ...string) string {
	toHash := strings.Join(ss, "-")
	hash := md5.Sum([]byte(toHash))
	return hex.EncodeToString(hash[:])
}

func getSecretFromVault(ctx context.Context, accessToken string, vaultUri, secretName, secretVersion string) (VaultSecret, error) {
	u := fmt.Sprintf("%s/secrets/%s", strings.TrimSuffix(vaultUri, "/"), secretName)
	if secretVersion != "" {
		u = fmt.Sprintf("%s/%s", u, secretVersion)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?api-version=7.3", u), nil)
	if err != nil {
		return VaultSecret{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return VaultSecret{}, err
	}

	if res.StatusCode != http.StatusOK {
		return VaultSecret{}, fmt.Errorf("unexpected status from vault api: %s", res.Status)
	}

	defer res.Body.Close()

	var v VaultSecret
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return VaultSecret{}, err
	}

	return v, nil
}
