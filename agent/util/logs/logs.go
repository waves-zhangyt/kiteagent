// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   util包中的日志工具，可以在程序层面控制输出级别。
*/
package logs

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
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
		print(debug, &v)
	}
}

func Info(v ...interface{}) {
	if *logLevel == "info" || *logLevel == "debug" {
		print(info, &v)
	}
}

func Warn(v ...interface{}) {
	if *logLevel == "info" || *logLevel == "warn" || *logLevel == "debug" {
		print(warn, &v)
	}
}

func Error(v ...interface{}) {
	print(error, &v)
}

// 统一打印格式
func print(logger *log.Logger, v interface{}) {
	vv := v.(*[]interface{})

	l := len(*vv)
	if l == 0 {
		logger.Println(getSourcePrefix())
	}

	if l > 1 {
		need, format := genFormat(vv)
		if need {
			logger.Printf(getSourcePrefix()+format+"\n", (*vv)[0:]...)
		} else {
			logger.Printf(getSourcePrefix()+fmt.Sprintf("%v", (*vv)[0])+"\n", (*vv)[1:]...)
		}

	} else {
		logger.Printf(getSourcePrefix() + fmt.Sprintf("%v", (*vv)[0]) + "\n")
	}
}

// 生成没有格式化字符串情况下的格式字符串
// 如果守参数是字符串，则默认为是格式字符串，不用生成
// 返回 (是否需要生成，生成的格式字符串)
func genFormat(v interface{}) (bool, string) {
	vv := v.(*[]interface{})

	if reflect.TypeOf((*vv)[0]).Name() != "string" {
		l := len(*vv)
		var builder strings.Builder

		for i := 0; i < l; i++ {
			builder.WriteString("%v")
			if i < l-1 {
				builder.WriteString(" ")
			}
		}
		return true, builder.String()
	}

	return false, ""
}
