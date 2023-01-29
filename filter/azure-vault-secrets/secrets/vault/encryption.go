package vault

import (
	"crypto/aes"
	"crypto/cipher"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Encryption struct {
	fs filesys.FileSystemOrOnDisk
}

func NewEnc(fs filesys.FileSystem) *Encryption {
	return &Encryption{fs: filesys.FileSystemOrOnDisk{FileSystem: fs}}
}

func (enc *Encryption) DecryptAtRest(path string, keyString string) ([]byte, error) {
	ciphertext, err := enc.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key := []byte(keyString)[:32]
	nonce := []byte(keyString)[len(keyString)-12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}

func (enc *Encryption) EncryptAtRest(path string, data []byte, keyString string) error {
	key := []byte(keyString)[:32]
	nonce := []byte(keyString)[len(keyString)-12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	return enc.fs.WriteFile(path, ciphertext)
}
