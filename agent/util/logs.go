// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   util包中的日志工具

   deprecated 20190330 周六 使用logs/logs.go中功能替代，这个不再推荐使用
*/
package util

import (
	"log"
	"os"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	Debug = log.New(os.Stdout, "Debug ", log.LstdFlags|log.Lshortfile)
	Info = log.New(os.Stdout, "Info ", log.LstdFlags|log.Lshortfile)
	Warning = log.New(os.Stdout, "Warning ", log.LstdFlags|log.Lshortfile)
	Error = log.New(os.Stderr, "Error ", log.LstdFlags|log.Lshortfile)
}
