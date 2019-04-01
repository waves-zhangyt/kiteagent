// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   httpproxy 包定内置http代理相关功能
*/
package httpproxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/waves-zhangyt/kiteagent/agent/cmd"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 处理http代理请求指令
func DoHttpProxy(command *cmd.Cmd) *cmd.CmdResult {
	var cmdResult cmd.CmdResult
	cmdResult.JobId = command.Head.JobId
	cmdResult.Type = cmd.ProxyHttp

	//允许超时
	timeout := command.Head.Timeout
	if timeout == 0 {
		timeout = cmd.DefaultCmdTimeout
	}

	//解析请求参数
	body := command.Body
	var request Request
	if json.Unmarshal([]byte(body), &request) != nil {
		logs.Error("解析请求参数出错")
		cmdResult.Stderr = "解析请求参数出错"
		return &cmdResult
	}

	var bodyData []byte
	bodyType := request.BodyType
	if bodyType == "BASE64" { //二进制内容
		bodyData, _ = base64.StdEncoding.DecodeString(request.Body)
	} else {
		bodyData = []byte(request.Body)
	}

	//执行代理请求
	responseCode, respHeaders, contextTypeHeadItems, respBody, errMsg := UniRequestWithResponseCode(request.Method, request.Url, request.Headers, bodyData, timeout)
	var response Response
	response.ResponseCode = responseCode
	response.Headers = respHeaders
	response.ContentType = contextTypeHeadItems[0]
	contentType := contextTypeHeadItems[0]

	//获取Content-Encoding头信息，因为可能目标返回的时压缩内容，这样的话，需要按照二进制数据处理
	contentEncoding := contextTypeHeadItems[1]

	//普通文本
	if (contentEncoding != "gzip") &&
		(strings.Contains(contentType, "xml") ||
			strings.Contains(contentType, "json") ||
			strings.Contains(contentType, "text") ||
			strings.Contains(contentType, "html")) {
		response.Body = string(respBody)
	} else {
		//二进制文件
		response.Body = base64.StdEncoding.EncodeToString(respBody)
	}
	jsonByte, _ := json.Marshal(response)
	cmdResult.Stdout = string(jsonByte)

	if errMsg != "" {
		cmdResult.Stderr = errMsg
	}

	return &cmdResult
}

type Response struct {
	ResponseCode int    `json:"responseCode"`
	Headers      string `json:"headers"`
	Body         string `json:"body"`
	ContentType  string `json:"contentType"`
}

// request请求数据结构（cmd.body）
type Request struct {
	Method   string             `json:"method"`
	Url      string             `json:"url"`
	Headers  *map[string]string `json:"headers"`
	Body     string             `json:"body"`
	BodyType string             `json:"bodyType"` //普通字符串还是base64编码的二进制内容字符串
	Timeout  int                `json:"timeout,omitempty"`
}

// 通用http请求 返回 （头信息的json字符串，[Content-Type, Content-Encoding]，body内容, err字符串）
func UniRequest(method, url string, headers *map[string]string, body []byte, timeout int) (string, []string, []byte,
	string) {
	_, header, headerContentType, body, err := UniRequestWithResponseCode(method, url, headers, body, timeout)
	return header, headerContentType, body, err
}

// 通用http请求 返回 （statusCode, 头信息的json字符串，[Content-Type, Content-Encoding]，body内容, err字符串）
func UniRequestWithResponseCode(method, url string, headers *map[string]string, body []byte, timeout int) (int, string, []string, []byte,
	string) {

	var client *http.Client
	if timeout <= 0 {
		client = &http.Client{
			//用cmd的默认超时时间
			Timeout: time.Duration(cmd.DefaultCmdTimeout) * time.Second,
		}
	} else {
		client = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	var buf io.Reader
	if body != nil {
		buf = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		logs.Error(err)
		return 500, "", []string{"", ""}, nil, err.Error()
	}

	if headers != nil {
		for k, v := range *headers {
			req.Header.Add(k, v)
		}
	}

	resp, err2 := client.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err2 != nil {
		logs.Error("http 请求错误 %s\n", err2)
		return 500, "", []string{"", ""}, nil, err2.Error()
	}

	//打印返回的头信息
	headJsonBytes, _ := json.MarshalIndent(resp.Header, "", " ")
	head := string(headJsonBytes)

	contentType := resp.Header.Get("Content-Type")
	contentEncoding := resp.Header.Get("Content-Encoding")

	//打印返回的body信息
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		logs.Error(err1)
		return 500, head, []string{contentType, contentEncoding}, nil, ""
	}

	return resp.StatusCode, head, []string{contentType, contentEncoding}, body, ""
}
