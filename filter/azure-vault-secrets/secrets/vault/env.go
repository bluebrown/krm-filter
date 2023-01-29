package vault

import (
	"errors"
	"os"
)

var ErrNoToken = errors.New("no token")

type EnvTokenReader struct {
	key string
}

func NewEnvTokenReader() EnvTokenReader {
	return EnvTokenReader{"AAD_ACCESS_TOKEN"}
}

func (e EnvTokenReader) AccessToken() (string, error) {
	if tok := os.Getenv(e.key); tok != "" {
		return tok, nil
	}
	return "", ErrNoToken
}

func (e EnvTokenReader) Salt() (string, error) {
	t, err := e.AccessToken()
	if err != nil {
		return "", err
	}
	return t[:32], nil
}
