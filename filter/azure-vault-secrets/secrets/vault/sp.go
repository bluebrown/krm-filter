package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	loginEndpoint = "https://login.microsoftonline.com/%s/oauth2/token"
	resource      = "https://vault.azure.net"
	scope         = "https://vault.azure.net/.default"
)

type ServicePrincipal struct {
	TenantId     string
	ClientId     string
	ClientSecret string
	confDir      string
	enc          *Encryption
}

type SpOption func(sp *ServicePrincipal)

func WithEncryption(enc *Encryption) SpOption {
	return func(sp *ServicePrincipal) {
		sp.enc = enc
	}
}

func NewServicePrincipalFromEnv(configDir string, opts ...SpOption) (ServicePrincipal, error) {
	sp := ServicePrincipal{confDir: configDir}
	for _, opt := range opts {
		opt(&sp)
	}

	if sp.enc == nil {
		sp.enc = new(Encryption)
	}

	sp.TenantId = os.Getenv("AZURE_TENANT_ID")
	sp.ClientId = os.Getenv("AAD_SERVICE_PRINCIPAL_CLIENT_ID")
	sp.ClientSecret = os.Getenv("AAD_SERVICE_PRINCIPAL_CLIENT_SECRET")

	if sp.TenantId == "" || sp.ClientId == "" || sp.ClientSecret == "" {
		return ServicePrincipal{}, errors.New("env vars not set")
	}

	// TODO: check is not very solid here. Since the sp is only used in other funcs
	// nothing stops the user from passing custom sp there
	if len(sp.ClientSecret) < 32 {
		return ServicePrincipal{}, errors.New("client secret must be minimum 32 bytes")
	}

	return sp, nil
}

func (sp ServicePrincipal) Salt() (string, error) {
	if len(sp.ClientSecret) < 32 {
		return "", errors.New("could not create salt")
	}
	return sp.ClientSecret[:32], nil
}

func (sp ServicePrincipal) AccessToken() (string, error) {

	// if the secret didn't exist in cache, get a token and fetch the secret
	// the is used from cache if it exists and is not expired
	var (
		tokenHash = getHash(sp.TenantId, sp.ClientId, sp.ClientSecret)
		tokenFile = path.Join(sp.confDir, tokenHash)
		refresh   = true
		token     = Token{}
	)

	cachedToken, err := sp.enc.DecryptAtRest(tokenFile, sp.ClientSecret)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("error while reading cached token: %w", err)
	}

	if !errors.Is(err, fs.ErrNotExist) {
		if err := json.Unmarshal(cachedToken, &token); err != nil {
			return "", fmt.Errorf("could not unmarshal cached token: %w", err)
		}
		if !isExpired(token) {
			refresh = false
		}
	}

	if refresh {
		t, err := login(context.TODO(), sp.TenantId, sp.ClientId, sp.ClientSecret)
		if err != nil {
			return "", err
		}
		token.AccessToken = t.AccessToken
		token.Expires, err = strconv.ParseInt(t.ExpiresOn, 10, 64)
		if err != nil {
			return "", err
		}

		b, err := json.Marshal(token)
		if err != nil {
			return "", fmt.Errorf("could not marshal token: %w", err)
		}
		err = sp.enc.EncryptAtRest(tokenFile, b, sp.ClientSecret)
		if err != nil {
			return "", fmt.Errorf("encrypt token: %w", err)
		}
	}

	return token.AccessToken, nil
}

func isExpired(token Token) bool {
	return time.Now().Unix() > token.Expires
}

func login(ctx context.Context, tenantId, clientId, clientSecret string) (TokenResponse, error) {
	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("client_id", clientId)
	data.Add("client_secret", clientSecret)
	data.Add("resource", resource)
	data.Add("scope", scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(loginEndpoint, tenantId), strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf("got wrong status code: %s", res.Status)
	}
	var v TokenResponse
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return TokenResponse{}, fmt.Errorf("failed to decode body: %w", err)
	}
	return v, nil
}
