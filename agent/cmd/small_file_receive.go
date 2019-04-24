package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"os"
)

// do receive small file
func ReceiveSmallFile(command *Cmd) *CmdResult {
	var cmdResult CmdResult
	cmdResult.JobId = command.Head.JobId
	cmdResult.Type = CmdSmallFileReceive

	//can timeout
	timeout := command.Head.Timeout
	if timeout == 0 {
		timeout = DefaultCmdTimeout
	}

	//parse the request body
	body := command.Body
	var smallFileData SmallFileData
	if json.Unmarshal([]byte(body), &smallFileData) != nil {
		logs.Error("解析请求参数出错")
		cmdResult.Stderr = "解析请求参数出错"
		return &cmdResult
	}

	// write the file
	var payloadBytes []byte
	var err error
	if smallFileData.Payload != "" {

		if smallFileData.PayloadType == "base64" {
			payloadBytes, err = base64.StdEncoding.DecodeString(smallFileData.Payload)
			if err != nil {
				errorMessage(&cmdResult, "decode the file payload failed: %v", err)
				return &cmdResult
			}
		} else {
			payloadBytes = []byte(smallFileData.Payload)
		}

		outFile, err := os.Create(smallFileData.TargetPath)
		if err != nil {
			errorMessage(&cmdResult, "decode the file payload failed: %v", err)
			return &cmdResult
		}
		defer outFile.Close()

		n, err := outFile.Write(payloadBytes)
		if err != nil {
			errorMessage(&cmdResult, "write data to the target file failed: %v", err)
			return &cmdResult
		}
		logs.Debug("%v bytes write to file %v", n, smallFileData.TargetPath)

		cmdResult.Stdout = "ok"
	}

	return &cmdResult

}

func errorMessage(cmdResult *CmdResult, msg string, err error) {
	text := fmt.Sprintf(msg, err)
	logs.Error(text)
	cmdResult.Stderr = text
}

type SmallFileData struct {
	TargetPath  string `json:"targetPath"`            // the target path
	PayloadType string `json:"payloadType,omitempty"` //normal string or base64
	Payload     string `json:"payload"`               // the real data in file, base64Encoded
}
