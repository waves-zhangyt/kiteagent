package httpserver

import (
	"github.com/waves-zhangyt/kiteagent/agent/conf"
	"github.com/waves-zhangyt/kiteagent/agent/util"
	"io"
	"log"
	"net/http"
	"strconv"
)

var Version string = "v0.5.0"

func InitServer() {
	go func() {
		// 版本问询服务
		http.HandleFunc("/version", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
			io.WriteString(writer, Version)
		})
		hostAndPort := ":" + strconv.Itoa(conf.DefaultConfig.HttpServerPort)
		util.Info.Printf("本地http监听: %s\n", hostAndPort)
		log.Fatal(http.ListenAndServe(hostAndPort, nil))
	}()
}
