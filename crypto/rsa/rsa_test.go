package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	bits := 2048
	prvKey, pubKey, _ := GenerateKeyPair(bits)
	prvKeyPEM := EncodeKeyToPEM(prvKey)
	pubKeyPEM := EncodeKeyToPEM(pubKey)
	text := "hello world"

	type args struct {
		text      string
		prvKeyPEM []byte
		pubKeyPEM []byte
		want      string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"encode-decode",
			args{
				text:      text,
				prvKeyPEM: prvKeyPEM,
				pubKeyPEM: pubKeyPEM,
				want:      text,
			},
			true,
		},
		{
			"empty text",
			args{
				text:      "",
				prvKeyPEM: prvKeyPEM,
				pubKeyPEM: pubKeyPEM,
				want:      "",
			},
			true,
		},
		{
			"empty Key",
			args{
				text:      text,
				prvKeyPEM: []byte{},
				pubKeyPEM: []byte{},
				want:      "",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			text2, err := EncryptString(tt.args.pubKeyPEM, tt.args.text)

			if err != nil && len(tt.args.pubKeyPEM) != 0 {
				t.Errorf("encryptString() error = %v", err)
				return
			}
			// 解密
			text3, err := DecryptString(tt.args.prvKeyPEM, text2)
			if err != nil && len(tt.args.prvKeyPEM) != 0 {
				t.Errorf("decryptString() error = %v", err)
				return
			}
			// 比较
			if text3 != tt.args.want {
				t.Errorf("RSA() = %v, want %v", text3, tt.args.want)
			}
			// pass
		})
	}
}

func TestSaveLoad(t *testing.T) {
	type args struct {
		text    string
		prvFile string
		pubFile string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"encode-decode1",
			args{
				text:    "hello world",
				prvFile: "test_prv.pem",
				pubFile: "test_pub.pem",
			},
			true,
		},
		{
			"encode-decode2",
			args{
				text:    "ddee",
				prvFile: "test_prv.pem",
				pubFile: "test_pub.pem",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits := 2048
			prvKey, pubKey, err := GenerateKeyPair(bits)
			if err != nil {
				t.Errorf("generateKeyPair() error = %v", err)
				return
			}
			// test prv key
			SaveKeyToFile(prvKey, tt.args.prvFile)
			prvKeyPEMFromFile, err := LoadKeyFromFile(tt.args.prvFile)
			if err != nil {
				t.Errorf("loadKeyFromFile() error = %v", err)
				return
			}
			prvKeyFromFile, err := DecodePrivateKeyFromPEM(prvKeyPEMFromFile)
			if err != nil {
				t.Errorf("decodePrivateKeyFromPEM() error = %v", err)
			} else if prvKeyFromFile.Equal(prvKey) != true {
				t.Errorf("RSA() = %v, want %v", prvKeyFromFile, prvKey)
			}
			// save pub key
			SaveKeyToFile(pubKey, tt.args.pubFile)
			// pass
		})
	}
}
