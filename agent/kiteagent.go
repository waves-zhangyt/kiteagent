// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   main包。 包含程序启动入口，初始化连接，命令分派等功能
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/waves-zhangyt/kiteagent/agent/cmd"
	"github.com/waves-zhangyt/kiteagent/agent/conf"
	"github.com/waves-zhangyt/kiteagent/agent/fork"
	"github.com/waves-zhangyt/kiteagent/agent/httpproxy"
	"github.com/waves-zhangyt/kiteagent/agent/httpserver"
	"github.com/waves-zhangyt/kiteagent/agent/util"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func askVersion() bool {
	args := flag.Args()
	if len(args) != 0 && args[0] == "version" {
		fmt.Println(httpserver.Version)
		return true
	}
	return false
}

var done = make(chan struct{})

// max try connect count 次数
const MaxTryConnectCount = 17280

var maxTryConnectCount = 0

// 异步队列，最多1000个累积
var asynCmdResultChannel = make(chan *cmd.CmdResult, 1000)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if !flag.Parsed() {
		flag.Parse()
	}

	// 版本命令执行
	if askVersion() {
		return
	}

	// 加载配置文件
	conf.SyncLoadConfig()

	// if start with daemon process
	if fork.Daemon() {
		return
	}

	// 初始化内置http服务
	httpserver.InitServer()

	//启动客户端，并进行命令分发
	var conn []*websocket.Conn
	conn = append(conn, nil)
	startWssClient(conn)

	//启动异步消息回写机制
	go asyncCmdResult(conn)

	//启动心跳
	go intervalPing(conn)

	//结束客户端执行入口
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-done:
			return
		case <-interrupt: //终结客户端后，发送关闭连接信息给服务端
			logs.Info("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			if conn[0] != nil {
				err := conn[0].WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					logs.Info("write close:", err)
					return
				}
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}

func connect(conn []*websocket.Conn) {
	//拨号
	defer func() {
		if e := recover(); e != nil {
			maxTryConnectCount++
			if maxTryConnectCount >= MaxTryConnectCount {
				logs.Error("重试连接次数超标，程序结束")
				os.Exit(1)
			}

			logs.Error("Panicing %s", e)
			logs.Info("等待5秒重新尝试连接")
			time.Sleep(5 * time.Second)
			connect(conn)
		}
	}()

	if conn[0] != nil { //从panic中恢复的时候，清除掉
		conn[0].Close()
	}

	dialer := websocket.Dialer{
		TLSClientConfig: createTLSConfig(), // tls 设置
	}

	url := conf.DefaultConfig.WssUrl + "?clientId=" + conf.DefaultConfig.AgentId + "&ipv4=" + util.LocalIPv4Addr() +
		"&connectionSecret=" + conf.DefaultConfig.ConnectionSecret
	connItem, _, err := dialer.Dial(url, nil)
	conn[0] = connItem
	if err != nil {
		msg := fmt.Sprintf("拨号失败: %s", err)
		panic(msg)
	} else {
		maxTryConnectCount = 0
	}
}

func createTLSConfig() *tls.Config {
	pool := x509.NewCertPool()
	caCertPath := conf.DefaultConfig.TlsPublicKey
	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		logs.Error("ReadFile err:", err)
		panic(err)
	}
	pool.AppendCertsFromPEM(caCrt)
	config := &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: true, //跳过x.509证书检测，使信任所有的证书（因为我用的是自签名证书，所以必须要有这个）
	}

	return config
}

func startWssClient(conn []*websocket.Conn) {
	//消息接收与派送给命令处理
	go func() {
		defer func() {
			if e := recover(); e != nil {
				maxTryConnectCount++
				if maxTryConnectCount >= MaxTryConnectCount {
					logs.Error("重试连接次数超标，程序结束")
					os.Exit(1)
				}

				logs.Error("Panicing %s", e)
				logs.Info("等待5秒重新尝试重新连接和读取")
				time.Sleep(5 * time.Second)
				startWssClient(conn)
			}
		}()
		connect(conn)
		for {
			//以json数据作为基础协议
			var v cmd.Cmd
			//读取json
			if conn[0] == nil {
				return
			}
			err := conn[0].ReadJSON(&v)
			if err != nil {
				errMsg := fmt.Sprintf("%s", err)
				if strings.Contains(errMsg, "close 1000 (normal)") {
					logs.Info("正常退出")
					close(done)
					conn[0] = nil
					return
				}

				logs.Error("read: %v", err)
				msg := fmt.Sprintf("读取错误 %s", err)
				conn[0] = nil
				panic(msg)
			} else {
				if maxTryConnectCount != 0 {
					maxTryConnectCount = 0
				}
			}
			//异步执行
			go func() {
				cmdResult := Dispatch(&v)

				if cmdResult != nil {
					logs.Debug("发送信息: %s", cmdResult)
				}

				if cmdResult == nil {
					//do nothing
				} else if v.Head.Async == 0 { // 同步消息
					if syncWriteJSon(conn[0], cmdResult) != nil {
						logs.Error("发送结果消息失败")
					}
				} else if v.Head.Async == 1 { // 异步消息
					//标志返回结果是异步的
					cmdResult.Async = 1
					//压入队列等待读取
					asynCmdResultChannel <- cmdResult
				}
			}()
		}
	}()
}

func asyncCmdResult(conn []*websocket.Conn) {
	for {
		cmdResult := <-asynCmdResultChannel
		if cmdResult != nil {
			err := syncWriteJSon(conn[0], cmdResult)
			if err != nil {
				logs.Error("写消息给客户端失败,重新压入原队列 %v", cmdResult)
				//发现错误从新返回队列
				asynCmdResultChannel <- cmdResult
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

//心跳检测
func intervalPing(conn []*websocket.Conn) {
	for {
		if conn[0] != nil {
			reqPing := cmd.CmdResult{
				Type:   cmd.Req_ping,
				Stdout: "ping",
			}
			err := syncWriteJSon(conn[0], reqPing)
			if err != nil {
				logs.Error("心跳检测失败: %s", err)
			}
		}
		//每隔一分钟一次心跳
		time.Sleep(time.Minute)
	}
}

//命令分发
func Dispatch(command *cmd.Cmd) *cmd.CmdResult {
	text, _ := json.Marshal(command)
	logs.Debug("收到信息：%s", text)

	cmdType := command.Head.Type

	if strings.HasPrefix(cmdType, cmd.Res_prefix) {
		return nil
	}

	switch cmdType {
	case cmd.CmdRun:
		return cmd.CommandRun(command)
	case cmd.ProxyHttp:
		return httpproxy.DoHttpProxy(command)
	default:
		logs.Info("不支持\"%s\"类型命令", cmdType)
		return nil
	}
}

// the websocket write mutex
var connWriteMutex sync.Mutex

// sync write for gorilla websocket
func syncWriteJSon(conn *websocket.Conn, v interface{}) error {
	connWriteMutex.Lock()
	defer connWriteMutex.Unlock()
	return conn.WriteJSON(v)
}
