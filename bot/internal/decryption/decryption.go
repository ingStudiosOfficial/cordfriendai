package decryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
)

func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("pkcs7: invalid padding size")
	}
	padLen := int(b[len(b)-1])
	if padLen == 0 || padLen > len(b) {
		return nil, fmt.Errorf("pkcs7: invalid padding")
	}
	for i := len(b) - padLen; i < len(b); i++ {
		if b[i] != byte(padLen) {
			return nil, fmt.Errorf("pkcs7: invalid padding")
		}
	}
	return b[:len(b)-padLen], nil
}

func DecryptAES256CBC(encryptedHex, ivHex string, key []byte) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("hex decode ciphertext: %w", err)
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", fmt.Errorf("hex decode iv: %w", err)
	}
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext not multiple of block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("unpad: %w", err)
	}
	return string(unpadded), nil
}
