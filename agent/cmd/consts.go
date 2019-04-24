// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   cmd基础包中定义的 常量
*/
package cmd

// 默认执行命令超时时间 单位 毫秒
const DefaultCmdTimeout int = 60000

// 返回结果前缀
const Res_prefix = "res_"

// 心跳消息
const Req_ping = "req_ping"

// 标准命令
const CmdRun = "cmd.run"

// http proxy
const ProxyHttp = "proxy.http"

// small file receive
// kite manager can transfer a small file like config file to the target kite agent host
const CmdSmallFileReceive = "cmd.smallFileReceive"
