package vault

type SecretFetchOptions struct {
	VaultUri      string
	SecretName    string
	SecretVersion string
}

type VaultSecret struct {
	Value      string                `json:"value"`
	ID         string                `json:"id"`
	Attributes VaultSecretAttributes `json:"attributes"`
	Tags       map[string]string     `json:"tags"`
}

type VaultSecretAttributes struct {
	Enabled         bool   `json:"enabled"`
	Created         int    `json:"created"`
	Updated         int    `json:"updated"`
	RecoveryLevel   string `json:"recoveryLevel"`
	RecoverableDays int    `json:"recoverableDays"`
}

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	Expires     int64  `json:"expires"`
}
