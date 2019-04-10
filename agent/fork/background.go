// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   provide background run method with -background=true parameter
*/
package fork

import (
	"flag"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"os"
	"os/exec"
	"time"
)

var background = flag.Bool("background", false, "run in background with -background=true")

func init() {

	if !flag.Parsed() {
		flag.Parse()
	}

	if *background {
		backgroundProcess()
	}

}

func backgroundProcess() {
	args := os.Args[1:] //omit the "background" parameter

	i := 0
	for ; i < len(args); i++ {
		if args[i] == "--background=true" || args[i] == "-background=true" || args[i] == "--background" ||
			args[i] == "-background" {
			args[i] = "-background=false"
			break
		}
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	logs.Info("[pid] %v", cmd.Process.Pid)

	time.Sleep(3 * time.Second)
	logs.Info("done")
	os.Exit(0)
}
