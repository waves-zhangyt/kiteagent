package cmd

import (
	"encoding/base64"
	"testing"
)

func TestReceiveSmallFile(t *testing.T) {

	var head Head
	head.JobId = "xxxx"
	head.Type = CmdSmallFileReceive

	var command Cmd
	command.Head = &head

	content := base64.StdEncoding.EncodeToString([]byte("你好，小文件1"))
	command.Body = "{\"targetPath\": \"smallFile.txt\",\"payloadType\":\"base64\", \"payload\": \"" + content + "\"}"

	cmdResult := ReceiveSmallFile(&command)
	if cmdResult.Stdout != "ok" {
		t.Errorf("recieve file err: %v", cmdResult.Stderr)
	}

	content = "普通文本"
	command.Body = "{\"targetPath\": \"smallFile1.txt\", \"payload\": \"" + content + "\"}"

	cmdResult = ReceiveSmallFile(&command)
	if cmdResult.Stdout != "ok" {
		t.Errorf("recieve file err: %v", cmdResult.Stderr)
	}

}
