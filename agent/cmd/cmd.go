// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE
/*
   cmd 包实现基础命令的数据模型，并实现基础命令业务逻辑
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/waves-zhangyt/kiteagent/agent/util"
	"os/exec"
	"runtime"
	"time"
)

// 基础命令 头信息结构
type Head struct {
	Type    string `json:"type"`              //like cmd.run
	JobId   string `json:"jobId,omitempty"`   //命令唯一序列号
	Timeout int    `json:"timeout,omitempty"` //执行命令超时时间 单位 毫秒
	Async   int    `json:"async,omitempty"`
}

// 基础命令结构
type Cmd struct {
	*Head `json:"head"`
	Body  string `json:"body,omitempty"` //实际的命令参数或二进制数据的base64编码
}

//基础命令 返回结果结构
type CmdResult struct {
	Type      string `json:"type,omitempty"`
	JobId     string `json:"jobId,omitempty"` //任务id, 命令唯一序列号
	Async     int    `json:"async"`           //标志是否异步命令
	IsTimeout bool   `json:"isTimeout,omitempty"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
}

func (cmdResult *CmdResult) String() string {
	data, err := json.Marshal(cmdResult)
	if err != nil {
		return fmt.Sprint(cmdResult)
	}
	return string(data)
}

// 执行通用命令
func CommandRun(cmd *Cmd) *CmdResult {
	var cmdResult CmdResult
	cmdResult.JobId = cmd.Head.JobId

	//允许命令超时
	timeout := cmd.Head.Timeout
	if timeout == 0 {
		timeout = DefaultCmdTimeout
	}
	//超时时间单位设定为秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer func() {
		if ctx.Err() != nil {
			cmdResult.IsTimeout = true
		}
		cancel()
	}()

	var theCmd *exec.Cmd //原来和我的参数同名 /(ㄒoㄒ)/~~，所以更改之
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		theCmd = exec.CommandContext(ctx, "bash", "-c", cmd.Body)
	} else {
		theCmd = exec.CommandContext(ctx, "cmd", "/c", cmd.Body)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	theCmd.Stdout = &stdout
	theCmd.Stderr = &stderr
	if err := theCmd.Run(); err != nil {
		//此种错误也输出到标准错误输出
		errMsg := fmt.Sprintf("%s", err)
		cmdResult.Stderr = errMsg
		util.Error.Println(errMsg)
	}

	cmdResult.Stdout = stringWithPlatform(stdout.Bytes())
	if cmdResult.Stderr != "" {
		cmdResult.Stderr = cmdResult.Stderr + " && " + stringWithPlatform(stderr.Bytes())
	} else {
		cmdResult.Stderr = stringWithPlatform(stderr.Bytes())
	}

	return &cmdResult
}

// 根据平台确定是否转换编码
// windows 默认用GBK, linux和mac默认用utf8
func stringWithPlatform(b []byte) string {
	if runtime.GOOS == "windows" {
		d, _ := util.GbkToUtf8(b)
		return string(d)
	} else {
		return string(b)
	}
}
