package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/scrypt"
)

const (
	DefaultStatePersistenceDelay = time.Second * 5
)

var (
	DefaultPassword = []byte("###multibase_storage_password###") // nolint: gochecknoglobals

	ErrNoData = errors.New("no data for decryption")
)

func Encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to prepare nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

func Decrypt(key, data []byte) ([]byte, error) {
	const saltLen = 32

	if len(data) <= saltLen {
		return nil, ErrNoData
	}

	salt, data := data[len(data)-saltLen:], data[:len(data)-saltLen]

	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gcm: %w", err)
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open gcm: %w", err)
	}

	return plaintext, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		const defaultSaltLen = 32

		salt = make([]byte, defaultSaltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, fmt.Errorf("failed to prepate salt: %w", err)
		}
	}

	// nolint: varnamelen
	const (
		n      = 4096
		r      = 8
		p      = 1
		keyLen = 32
	)

	key, err := scrypt.Key(password, salt, n, r, p, keyLen)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate scrypt key: %w", err)
	}

	return key, salt, nil
}
