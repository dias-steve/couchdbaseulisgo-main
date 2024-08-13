package couchdbUtils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// Fonction pour chiffrer des données
func Encrypt(data, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	return fmt.Sprintf("%x:%x", nonce, ciphertext), nil
}

// Fonction pour déchiffrer des données
func Decrypt(encryptedData string, key []byte) (string, error) {
	parts := strings.Split(encryptedData, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted data format")
	}

	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
