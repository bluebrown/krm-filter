package filemock

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Source struct {
	srcDir string
}

func NewSource(srcDir string) Source {
	p, err := filepath.Abs(srcDir)
	if err != nil {
		panic(err)
	}
	return Source{p}
}

func (src Source) RetrieveSecret(uri, name, version string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	vaultName := strings.Split(u.Host, ".")[0]
	fname := fmt.Sprintf("%s.%s", vaultName, name)
	if version != "" {
		fname = fmt.Sprintf("%s.%s", fname, version)
	}
	b, err := os.ReadFile(filepath.Join(src.srcDir, fname))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
