// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   provide a http server for app version
*/
package httpserver

import (
	"github.com/waves-zhangyt/kiteagent/agent/conf"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

var Version string = "v0.7.0"

func InitServer() {
	go func() {
		// 版本问询服务
		http.HandleFunc("/version", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
			io.WriteString(writer, Version)
		})
		hostAndPort := ":" + strconv.Itoa(conf.DefaultConfig.HttpServerPort)
		logs.Info("本地http监听: %s", hostAndPort)
		log.Fatal(http.ListenAndServe(hostAndPort, nil))
	}()
}
