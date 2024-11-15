package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	conf string
)

type NacosConfig struct {
	IP       string
	Port     int64
	Username string
	Password string
}

func init() {
	flag.StringVar(&conf, "conf", "", "config path, eg: -conf config.yaml")
}

func loadConfig() NacosConfig {
	// 读取配置文件
	configData, err := os.ReadFile(conf)
	if err != nil {
		panic(err)
	}

	var conf NacosConfig
	if err := json.Unmarshal(configData, &conf); err != nil {
		panic(err)
	}
	return conf
}

func main() {
	flag.Parse()
	data := loadConfig()

	// 创建 HTTP 处理函数
	http.HandleFunc("/nacos_config", func(w http.ResponseWriter, r *http.Request) {
		response := NacosConfig{
			IP:       data.IP,
			Port:     data.Port,
			Username: data.Username,
			Password: data.Password, // 秘钥在对应解密的项目中
		}

		// 编码为 JSON 格式
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "JSON encoding error", http.StatusInternalServerError)
			return
		}

		// 设置响应头，指定响应类型为 JSON
		w.Header().Set("Content-Type", "application/json")

		// 写入 JSON 数据到响应
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, "Response writing error", http.StatusInternalServerError)
			return
		}
	})

	// 指定监听的端口
	port := "13579"

	// 启动 HTTP 服务器
	fmt.Printf("Server is running on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
