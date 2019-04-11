// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   conf包主要定义配置及解析
*/
package conf

import (
	"flag"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sync"
)

// 基本配置结构
type Config struct {
	AgentId        string `json:"agentId" yaml:"agentId"`               //客户端唯一标志
	WssUrl         string `json:"wssUrl" yaml:"wssUrl"`                 //连接服务端url
	TlsPublicKey   string `json:"tlsPublicKey" yaml:"tlsPublicKey"`     //tls 自签名证书文件路径
	HttpServerPort int    `json:"httpServerPort" yaml:"httpServerPort"` //内置http服务端口
}

// 默认配置变量
var DefaultConfig Config
var configLoaded bool
var configLoadedMu sync.Mutex

var agentId = flag.String("a", "", "agentId")
var wssUrl = flag.String("m", "", "wssUrl")
var configFile = flag.String("c", "conf/kite-agent.yml", "kite-agent.yml file path") //默认相对路径
var tlsPublicKey = flag.String("tls", "", "tls public key file path")

// 判断路径是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	return false
}

// load config multi goruntine safe
func SyncLoadConfig() {
	configLoadedMu.Lock()
	defer configLoadedMu.Unlock()

	if !configLoaded {
		LoadConfig()
		configLoaded = true
	}
}

// 加载配置
func LoadConfig() *Config {

	if !flag.Parsed() {
		flag.Parse()
	}

	//1.读取配置文件，如果和1中的有重叠，则覆盖
	if pathExists(*configFile) {
		data, err := ioutil.ReadFile(*configFile)
		if err != nil {
			panic(err)
		}
		if err := yaml.Unmarshal(data, &DefaultConfig); err != nil {
			logs.Info("yaml unmarshaling failed: %s", err)
		}
	}

	//2.命令行读取。如果命令行中存在，则应用命令中的，优先级最高
	if *agentId != "" {
		DefaultConfig.AgentId = *agentId
	}
	if *wssUrl != "" {
		DefaultConfig.WssUrl = *wssUrl
	}
	if *tlsPublicKey != "" {
		DefaultConfig.TlsPublicKey = *tlsPublicKey
	}

	//3.未设置部分有些从操作系统获取
	if DefaultConfig.AgentId == "" {
		DefaultConfig.AgentId, _ = os.Hostname()
	}

	rdata, _ := yaml.Marshal(DefaultConfig)
	logs.Info("当前应用配置信息\n-------\n%s-------~", string(rdata))

	return &DefaultConfig
}
