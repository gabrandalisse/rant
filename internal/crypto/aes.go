package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const MasterKeySize = 32

var ErrInvalidKeySize = errors.New("master key must be 32 bytes")

type AESGCMCipher struct {
	aead cipher.AEAD
}

func NewAESGCMCipher(masterKey []byte) (*AESGCMCipher, error) {
	if len(masterKey) != MasterKeySize {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESGCMCipher{aead: aead}, nil
}

func (c *AESGCMCipher) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := c.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (c *AESGCMCipher) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	return c.aead.Open(nil, nonce, ciphertext, nil)
}
