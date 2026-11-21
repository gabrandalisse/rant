package keyring

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	zkeyring "github.com/zalando/go-keyring"
)

const (
	serviceName = "rant"
	accountName = "master-key"
	keyByteSize = 32
)

type OSKeyStore struct{}

func NewOSKeyStore() *OSKeyStore {
	return &OSKeyStore{}
}

func (k *OSKeyStore) GetOrCreateMasterKey() ([]byte, error) {
	encoded, err := zkeyring.Get(serviceName, accountName)
	if err == nil {
		key, decodeErr := hex.DecodeString(encoded)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode master key: %w", decodeErr)
		}
		return key, nil
	}

	if !errors.Is(err, zkeyring.ErrNotFound) {
		return nil, fmt.Errorf("read keyring: %w", err)
	}

	rawKey := make([]byte, keyByteSize)
	if _, err := rand.Read(rawKey); err != nil {
		return nil, fmt.Errorf("generate master key: %w", err)
	}

	if err := zkeyring.Set(serviceName, accountName, hex.EncodeToString(rawKey)); err != nil {
		return nil, fmt.Errorf("store master key in keyring: %w", err)
	}

	return rawKey, nil
}

func (k *OSKeyStore) DeleteMasterKey() error {
	err := zkeyring.Delete(serviceName, accountName)
	if errors.Is(err, zkeyring.ErrNotFound) {
		return nil
	}
	return err
}
