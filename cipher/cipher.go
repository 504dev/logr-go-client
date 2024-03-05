package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

func EncryptAesJson(data interface{}, priv string) ([]byte, error) {
	privBytes, _ := base64.StdEncoding.DecodeString(priv)
	jsonMsg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return EncryptAes(jsonMsg, privBytes)
}

func DecodeAesJson(cipherBytes []byte, priv string, dest interface{}) error {
	privBytes, _ := base64.StdEncoding.DecodeString(priv)
	text, err := DecryptAes(cipherBytes, privBytes)
	if err != nil {
		return err
	}
	err = json.Unmarshal(text, dest)
	if err != nil {
		return err
	}
	return nil
}

func EncryptAes(plainText []byte, key []byte) ([]byte, error) {
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return cipherText, err
}

func EncryptAesIv(plainText []byte, key []byte, iv []byte) ([]byte, error) {
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	salt := sha256.Sum256(iv)
	iv = salt[:aes.BlockSize]
	cipherText := append(iv, make([]byte, len(plainText))...)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return cipherText, err
}

func DecryptAes(cipherText []byte, key []byte) ([]byte, error) {
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return nil, err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
