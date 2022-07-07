package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

func genMasterKey(password string, email string) []byte {
	return pbkdf2.Key([]byte(password), []byte(email), 100000, 32, sha256.New)
}

func strechMasterKey(masterKey []byte) ([]byte, []byte) {
	hkdfKey := make([]byte, 32)
	hkdfMacKey := make([]byte, 32)

	reader := hkdf.Expand(sha256.New, masterKey, []byte("enc"))
	reader.Read(hkdfKey)

	reader = hkdf.Expand(sha256.New, masterKey, []byte("mac"))
	reader.Read(hkdfMacKey)

	return hkdfKey, hkdfMacKey
}

func genProtectedSymKey(key []byte, macKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	symKey := make([]byte, 64)
	_, err = rand.Read(symKey)
	if err != nil {
		return nil, err
	}

	protectedSymKey := make([]byte, 64)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(protectedSymKey, symKey)
	ciphertext := append(iv, protectedSymKey...)

	mac := hmac.New(sha256.New, macKey)
	mac.Write(ciphertext)

	return mac.Sum(ciphertext), nil
}

func decryptProtectedSymKey(key []byte, macKey []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, macKey)
	mac.Write(ciphertext[:len(ciphertext)-mac.Size()])
	if !hmac.Equal(ciphertext[len(ciphertext)-mac.Size():], mac.Sum(nil)) {
		return nil, errors.New("MACs don't match")
	}

	iv := ciphertext[:aes.BlockSize]
	decryptedSymKey := make([]byte, 64)
	cbcDecrypter := cipher.NewCBCDecrypter(block, iv)
	cbcDecrypter.CryptBlocks(decryptedSymKey, ciphertext[aes.BlockSize:len(ciphertext)-mac.Size()])

	return decryptedSymKey, nil
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}

	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}

	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}

	return data[:length-padLen], nil
}

func padPKCS7(data []byte, blockSize int) ([]byte, error) {
	if blockSize < 0 || blockSize > 256 {
		return nil, fmt.Errorf("pkcs7: Invalid block size %d", blockSize)
	}

	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...), nil
}
