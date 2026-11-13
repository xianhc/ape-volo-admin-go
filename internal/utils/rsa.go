package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
)

// rsaKeyBits 密码传输要求的密钥长度。
//
// 4096 位对应 512 字节密文，Base64 后固定 684 字符，
// 前端按这个长度做入参校验，密钥规格不一致会直接对不上。
const rsaKeyBits = 4096

var (
	rsaKeyOnce sync.Once
	rsaKey     *rsa.PrivateKey
	rsaKeyErr  error
)

// loadRsaPrivateKey 读取并解析私钥，只做一次。
//
// 解析失败后不再重试：配置错了应该在启动时就暴露，而不是每个请求重试一遍。
func loadRsaPrivateKey() (*rsa.PrivateKey, error) {
	rsaKeyOnce.Do(func() {
		pemBytes, err := readRsaPrivateKeyPem()
		if err != nil {
			rsaKeyErr = err
			return
		}
		rsaKey, rsaKeyErr = parseRsaPrivateKey(pemBytes)
	})
	return rsaKey, rsaKeyErr
}

func readRsaPrivateKeyPem() ([]byte, error) {
	// 内联内容优先，便于容器用环境变量注入而不落盘
	if inline := global.Config.Rsa.PrivateKey; inline != "" {
		return []byte(inline), nil
	}

	path := global.Config.Rsa.PrivateKeyPath
	if path == "" {
		return nil, errors.New("未配置 RSA 私钥：请设置 rsa.private-key-path 或 rsa.private-key")
	}

	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 RSA 私钥失败(%s): %w", path, err)
	}
	return pemBytes, nil
}

func parseRsaPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("RSA 私钥不是合法的 PEM 格式")
	}

	// PKCS#8 是通用格式；兼容 PKCS#1 便于本地用 openssl genrsa 直接生成
	var key *rsa.PrivateKey
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		typed, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("RSA 私钥类型不匹配")
		}
		key = typed
	} else {
		legacy, legacyErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if legacyErr != nil {
			return nil, fmt.Errorf("解析 RSA 私钥失败: %w", err)
		}
		key = legacy
	}

	if key.N.BitLen() != rsaKeyBits {
		return nil, fmt.Errorf("RSA 私钥必须为 %d 位，当前为 %d 位", rsaKeyBits, key.N.BitLen())
	}
	return key, nil
}

// ValidateRsaKey 校验私钥可用，供启动时快速失败。
func ValidateRsaKey() error {
	_, err := loadRsaPrivateKey()
	return err
}

// Decrypt 用私钥解密前端传来的密码密文。
//
// 密文必须是 Base64 编码的 RSA-OAEP-SHA256 结果。
//
// # 为什么用 OAEP 而不是 PKCS#1 v1.5
//
// PKCS#1 v1.5 在解密失败时会产生可区分的错误行为，攻击者能据此判断
// 填充是否合法，逐步还原明文（Bleichenbacher 填充预言攻击）。
// OAEP 的解密过程不泄露这类信息。
//
// # 失败场景
//
// 私钥未配置、密文不是合法 Base64、密文与密钥不匹配时返回错误。
func Decrypt(cipherText string) (string, error) {
	if cipherText == "" {
		return "", errors.New("密文为空")
	}

	key, err := loadRsaPrivateKey()
	if err != nil {
		return "", err
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("密文不是合法的 Base64: %w", err)
	}

	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, cipherBytes, nil)
	if err != nil {
		return "", fmt.Errorf("RSA 解密失败: %w", err)
	}
	return string(plain), nil
}

// Encrypt 用私钥推导出的公钥加密，供本地联调与测试使用。
//
// 生产链路由前端加密，服务端只负责解密。
func Encrypt(plainText string) (string, error) {
	key, err := loadRsaPrivateKey()
	if err != nil {
		return "", err
	}

	// OAEP-SHA256 的明文上限为 密钥字节数 - 2*摘要长度 - 2
	maxLen := key.Size() - 2*sha256.Size - 2
	if len(plainText) > maxLen {
		return "", fmt.Errorf("明文长度 %d 超过 RSA-OAEP 上限 %d 字节", len(plainText), maxLen)
	}

	cipherBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, []byte(plainText), nil)
	if err != nil {
		return "", fmt.Errorf("RSA 加密失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}
