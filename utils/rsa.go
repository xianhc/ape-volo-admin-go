package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"go-apevolo/global"
)

func Decrypt(cipherText string) (plaintext string, err error) {
	// Base64 解码密文
	ciphertextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return
	}

	// 解码 DER 编码的私钥数据
	privateKeyData, _ := base64.StdEncoding.DecodeString(global.Config.Rsa.PrivateKey)

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyData)
	if err != nil {
		return
	}

	// 使用私钥解密数据
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertextBytes)
	if err != nil {
		return
	}

	return string(plain), err
}
