package config

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	nacos "github.com/go-kratos/kratos/contrib/config/nacos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"io"
	"net/http"
)

var (
	conf             string
	nacosProxy       string
	nacosNamespaceID string
	nacosDataID      string
	nacosGroup       string
)

type NacosConfig struct {
	IP       string
	Port     int64
	Username string
	Password string
}

func init() {
	flag.StringVar(&conf, "conf", "", "config path, eg: -conf config.yaml")
	flag.StringVar(&nacosProxy, "nacos-proxy", "", "nacos proxy")
	flag.StringVar(&nacosNamespaceID, "nacos-namespace-id", "", "nacos namespace id")
	flag.StringVar(&nacosDataID, "nacos-data-id", "", "nacos data id")
	flag.StringVar(&nacosGroup, "nacos-group", "", "nacos group")
}

func LoadConf(prefix string, bootstrap any) error {
	flag.Parse()

	var source config.Source
	switch {
	case conf != "":
		source = file.NewSource(conf)
	case nacosProxy != "":
		source = newNacosSource()
	default:
		conf = "../../configs"
		source = file.NewSource(conf)
	}

	c := config.New(
		config.WithSource(
			env.NewSource(prefix),
			source,
		),
	)
	if err := c.Load(); err != nil {
		return err
	}
	if err := c.Scan(bootstrap); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(bootstrap, "", "  ")
	if err != nil {
		return err
	}
	log.Debug("config:\n", string(raw))
	return nil
}

func newNacosSource() config.Source {
	var nacosConfig NacosConfig
	// 发请求 拿到nacos配置 /ias-kit/config/nacos/main.go
	remoteUrl := nacosProxy + "/nacos_config"
	response, err := DoGet(remoteUrl)
	if err != nil {
		panic(err)
	}
	// 解析返回参数
	err = json.Unmarshal(response, &nacosConfig)
	if err != nil {
		panic(err)
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nacosConfig.IP, uint64(nacosConfig.Port)),
	}

	cc := &constant.ClientConfig{
		NamespaceId:         nacosNamespaceID,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}
	if nacosConfig.Username != "" && nacosConfig.Password != "" {
		cc.Username = nacosConfig.Username
		// 解析密码
		cipherTextBytes, err := hex.DecodeString(nacosConfig.Password)
		if err != nil {
			panic(err)
		}
		decrypted, err := decrypt(cipherTextBytes)
		if err != nil {
			panic(err)
		}
		cc.Password = string(decrypted)
	}

	// a more graceful way to create naming client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	return nacos.NewConfigSource(
		client,
		nacos.WithGroup(nacosGroup),
		nacos.WithDataID(nacosDataID),
	)
}

func DoGet(url string) ([]byte, error) {
	client := &http.Client{}
	// 构建GET请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte(""), err
	}
	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	// 读取响应的内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}

// nacos密码解密
func decrypt(cipherText []byte) ([]byte, error) {
	// 密钥
	key := []byte("${sk}")
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf("cipherText is too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
