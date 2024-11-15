package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	IAS_PUB_TYPE = "IAS_PUB_KEY"
	IAS_PRV_TYPE = "IAS_PRV_KEY"
)

// 使用RSA公钥加密字符串，返回Base64编码的密文
func EncryptString(publicKeyPEM []byte, plaintext string) (string, error) {
	// 解码公钥PEM格式
	rsaPublicKey, err := DecodePublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", err
	}

	// 使用公钥加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("加密失败: %v", err)
	}

	// Base64编码
	encodedStr := base64.StdEncoding.EncodeToString(ciphertext)

	return encodedStr, nil
}

// 使用RSA私钥解密字符串
func DecryptString(privateKeyPEM []byte, ciphertext string) (string, error) {
	// Base64解码
	decodedBytes, err := base64.StdEncoding.DecodeString(ciphertext)

	// 解码私钥PEM格式
	privateKey, err := DecodePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}

	// 使用私钥解密
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decodedBytes)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

// 生成密钥对
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成私钥失败: %v", err)
	}

	// 生成公钥
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

// rsa格式的密钥转换为pem格式
func EncodeKeyToPEM(key interface{}) []byte {
	switch key.(type) {
	case *rsa.PrivateKey:
		keyBytes := x509.MarshalPKCS1PrivateKey(key.(*rsa.PrivateKey))
		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  IAS_PRV_TYPE,
			Bytes: keyBytes,
		})
		return keyPEM
	case *rsa.PublicKey:
		keyBytes, _ := x509.MarshalPKIXPublicKey(key.(*rsa.PublicKey))
		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  IAS_PUB_TYPE,
			Bytes: keyBytes,
		})
		return keyPEM
	default:
		return nil
	}
}

// pem格式的私钥转换为rsa格式
func DecodePrivateKeyFromPEM(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	// 解码私钥PEM格式
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != IAS_PRV_TYPE {
		return nil, fmt.Errorf("无效的私钥")
	}

	// 解析私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}
	return privateKey, nil
}

// pem格式的公钥转换为rsa格式
func DecodePublicKeyFromPEM(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	// 解码公钥PEM格式
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != IAS_PUB_TYPE {
		return nil, fmt.Errorf("无效的公钥")
	}

	// 解析公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %v", err)
	}

	// 类型断言为RSA公钥
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("无效的RSA公钥")
	}

	return rsaPublicKey, nil
}

// 将生成的密钥保存至本地文件
func SaveKeyToFile(key interface{}, filename string) error {
	// 创建密钥文件
	keyFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建密钥文件失败: %v", err)
	}
	defer keyFile.Close()

	// 写入密钥
	switch key.(type) {
	case *rsa.PrivateKey:
		privateKeyPEM := &pem.Block{
			Type:  IAS_PRV_TYPE,
			Bytes: x509.MarshalPKCS1PrivateKey(key.(*rsa.PrivateKey)),
		}
		err = pem.Encode(keyFile, privateKeyPEM)
		if err != nil {
			return fmt.Errorf("写入私钥文件失败: %v", err)
		}
	case *rsa.PublicKey:
		publicKeyPEM := &pem.Block{
			Type:  IAS_PUB_TYPE,
			Bytes: x509.MarshalPKCS1PublicKey(key.(*rsa.PublicKey)),
		}
		err = pem.Encode(keyFile, publicKeyPEM)
		if err != nil {
			return fmt.Errorf("写入公钥文件失败: %v", err)
		}
	default:
		return fmt.Errorf("无效的密钥类型")
	}
	return nil
}

// 从本地读取密钥文件
func LoadKeyFromFile(filename string) ([]byte, error) {
	// 打开密钥文件
	keyFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开密钥文件失败: %v", err)
	}
	defer keyFile.Close()

	// 读取密钥文件
	keyFileStat, err := keyFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("读取密钥文件失败: %v", err)
	}
	keyFileBytes := make([]byte, keyFileStat.Size())
	_, err = keyFile.Read(keyFileBytes)
	if err != nil {
		return nil, fmt.Errorf("读取密钥文件失败: %v", err)
	}

	return keyFileBytes, nil
}
