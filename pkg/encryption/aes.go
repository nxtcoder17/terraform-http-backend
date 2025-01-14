package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type AESCipher struct {
	Cipher cipher.Block
}

func NewAESCipher(key []byte) (Cipher, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &AESCipher{Cipher: cipher}, nil
}

// Encrypt implements Cipher.
func (ac *AESCipher) Encrypt(b []byte) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(ac.Cipher, iv)
	stream.XORKeyStream(ciphertext[ac.Cipher.BlockSize():], b)

	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(ciphertext)))

	base64.RawStdEncoding.Encode(encoded, ciphertext)
	return encoded, nil
}

// Decrypt implements Cipher.
func (ac *AESCipher) Decrypt(b []byte) ([]byte, error) {
	ciphertext := make([]byte, base64.RawStdEncoding.DecodedLen(len(b)))
	if _, err := base64.RawStdEncoding.Decode(ciphertext, b); err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:ac.Cipher.BlockSize()]
	ciphertext = ciphertext[ac.Cipher.BlockSize():]

	stream := cipher.NewCFBDecrypter(ac.Cipher, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	// Trim padding
	ciphertext = bytes.TrimRight(ciphertext, "\x00")
	return ciphertext, nil
}
