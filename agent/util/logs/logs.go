// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   util包中的日志工具，可以在程序层面控制输出级别。
*/
package logs

import (
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
)

var logLevel = flag.String("logLevel", "debug", "log level: debug info warn error")

func init() {
	debug = log.New(os.Stdout, "Debug ", log.LstdFlags)
	info = log.New(os.Stdout, "Info ", log.LstdFlags)
	warn = log.New(os.Stdout, "Warn ", log.LstdFlags)
	error = log.New(os.Stderr, "Error ", log.LstdFlags)
}

func getSourcePrefix() string {
	_, file, line, ok := runtime.Caller(2)

	refile := file
	if ok {
		//fmt.Println("Func Name=" + runtime.FuncForPC(funcName).Name())
		idx := -1
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			idx = strings.LastIndex(file, "/")
		} else {
			idx = strings.LastIndex(file, "\\")
		}
		if idx != -1 {
			refile = file[idx+1:]
		}

		return refile + ":" + strconv.Itoa(line) + " "
	}

	return ""
}

func Debug(v ...interface{}) {
	if *logLevel == "debug" {
		if len(v) > 1 {
			debug.Printf(getSourcePrefix()+v[0].(string)+"\n", v[1:]...)
		} else {
			debug.Printf(getSourcePrefix() + v[0].(string) + "\n")
		}
	}
}

func Info(v ...interface{}) {
	if *logLevel == "info" || *logLevel == "debug" {
		if len(v) > 1 {
			info.Printf(getSourcePrefix()+v[0].(string)+"\n", v[1:]...)
		} else {
			info.Printf(getSourcePrefix() + v[0].(string) + "\n")
		}
	}
}

func Warn(v ...interface{}) {
	if *logLevel == "info" || *logLevel == "warn" || *logLevel == "debug" {
		if len(v) > 1 {
			warn.Printf(getSourcePrefix()+v[0].(string)+"\n", v[1:]...)
		} else {
			warn.Printf(getSourcePrefix() + v[0].(string) + "\n")
		}

	}
}

func Error(v ...interface{}) {
	if len(v) > 1 {
		error.Printf(getSourcePrefix()+v[0].(string)+"\n", v[1:]...)
	} else {
		error.Printf(getSourcePrefix() + v[0].(string) + "\n")
	}

}
