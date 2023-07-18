package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Passwd    string  `json:"passwd"`
	PublicKey string  `json:"publicKey"`
	User      string  `json:"user"`
	IsSecret  bool    `json:"issecret"`
	Port      int     `json:"port"`
	Hosts     []Hosts `json:"hosts"`
}

type Hosts struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Passwd    string `json:"passwd"`
	PublicKey string `json:"publicKey"`
}

func Init() *Config {
	// 读取配置文件
	data, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal("读取config配置文件失败: ", err)
	}

	// 解析json格式配置文件
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("解码json数据失败: ", err)
	}

	return &config
}
