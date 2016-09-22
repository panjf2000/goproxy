// Package config provides Config struct for proxy.
package config

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// Config 保存代理服务器的配置
type Config struct {
	// 代理服务器工作端口,eg:":8080"
	Port string `json:"port"`

	// 反向代理标志
	Reverse bool `json:"reverse"`

	// 反向代理目标地址,eg:"127.0.0.1:8080"
	ProxyPass []string `json:"proxy_pass"`

	// 负载策略 1-IP HASH , 2-随机择取
	Mode int
	// 认证标志
	Auth bool `json:"auth"`

	// 缓存标志
	Cache bool `json:"cache"`

	// redis
	RedisHost string `json:"redis_host"`

	// redis passwd
	RedisPasswd string `json:"redis_passwd"`

	// 缓存定期刷新时间，单位分钟
	CacheTimeout int64 `json:"cache_timeout"`

	// 日志信息，1输出Debug信息，0输出普通监控信息
	Log int `json:"log"`

	// 管理员账号
	Admin map[string]string `json:"admin"`
	// 普通用户账户
	User map[string]string `json:"user"`
}

// GetConfig gets config from json file.
// GetConfig 从指定json文件读取config配置
func (c *Config) GetConfig(filename string) error {
	c.Admin = make(map[string]string)
	c.User = make(map[string]string)
	c.ProxyPass = make([]string, 10, 100)

	configFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer configFile.Close()

	br := bufio.NewReader(configFile)
	err = json.NewDecoder(br).Decode(c)
	if err != nil {
		return err
	}
	return nil
}

// WriteTOFile writes config into json file.
// WriteToFile 将config配置写入特定json文件
func (c *Config) WriteToFile(filename string) error {
	configFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer configFile.Close()

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	cjson := string(b)
	cspilts := strings.Split(cjson, ",")
	cjson = strings.Join(cspilts, ",\n")

	configFile.Write([]byte(cjson))

	return nil
}
