package main

import (
	"github.com/waves-zhangyt/kiteagent/agent/cmd"
	"testing"
)

func TestIsOutOfMaxThreshold(t *testing.T) {

	var result cmd.CmdResult
	result.Stderr = "hello"
	result.Stdout = "世界,i write the data"

	if IsOutOfMaxThreshold(&result) {
		t.Errorf("IsOutOfMaxThreshold(%v)=true", result)
	}

}
