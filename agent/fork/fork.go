// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   provide the monitor function for lang alive process
*/
package fork

import (
	"flag"
	"github.com/waves-zhangyt/kiteagent/agent/conf"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// child process pid
var childPid int

// daemon provide a master process to monitor the work process. if you want run app in backgroud use nohup in *linux or
// start /b in windows or use other methods you familiar with
var daemon = flag.Bool("daemon", false, "if start with a daemon process")
var logFile = flag.String("logFile", "kiteagent.log", "log file path when start with -daemon")
var healthErrCnt int

// check if has -daemon parameter and fork a process when it's true
func Daemon() bool {

	// run daemon
	if *daemon {
		// fork the work process
		fork()

		go healthCheck()

		// logfile cutter when worker process log file is to large, now the threshold is 512M
		go logFileCutterCheck(*logFile)

		//terminate entrance for the parent process
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		for {
			select {
			case <-interrupt:
				logs.Info("parent process interrupted normally")
				// note: undo, when in windows normal shut parent process, the child process killed failed
				killChildProcess()
				return true
			}
		}
	}

	return false
}

// log file cutter check
// may optimize later
func logFileCutterCheck(logFile string) {

	for {
		time.Sleep(5 * time.Minute)

		fileSize := getFileSize(logFile)
		if fileSize >= 1024*1024*512 {
			cutCmd := "echo '' > " + logFile
			var theCmd *exec.Cmd
			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				theCmd = exec.Command("bash", "-c", cutCmd)
			} else {
				theCmd = exec.Command("cmd", "/c", cutCmd)
			}

			err := theCmd.Run()
			if err != nil {
				logs.Error("cut the logfile %v err %v", logFile, err)
			}
		}
	}

}

// get file size
func getFileSize(fileName string) int64 {
	fi, err := os.Lstat(fileName)
	if err != nil {
		logs.Error("get the file %v info err %v", fileName, err)
	}
	return fi.Size()
}

// health check every 5 seconds
// if err count >= 3 times, fork a new child process
func healthCheck() {
	for {
		time.Sleep(5 * time.Second)

		if IsUpdating() {
			logs.Info("health check jumped for updating")
			continue
		}

		version := httpGet("http://127.0.0.1:" + strconv.Itoa(conf.DefaultConfig.HttpServerPort) + "/version")
		if !isActive(string(version)) {
			healthErrCnt++
		} else {
			healthErrCnt = 0
		}

		if healthErrCnt >= 3 {
			logs.Warn("check child process unhealthy, start a new one")
			killChildProcess()
			fork()
		}
	}
}

func isActive(version string) bool {
	if strings.HasPrefix(version, "v") {
		return true
	}

	return false
}

var cmd *exec.Cmd

// start new process as child process
func fork() {
	args := os.Args[1:]
	i := 0
	for ; i < len(args); i++ {
		if args[i] == "-daemon=true" || args[i] == "-daemon" {
			args[i] = "-daemon=false"
			break
		}
	}

	logs.Info("fork process cmd is %s", os.Args[0])
	cmd = exec.Command(os.Args[0], args...)
	// forked process log
	outFile, err := os.OpenFile(*logFile, os.O_RDWR|os.O_APPEND, 0)
	if err != nil && os.IsNotExist(err) {
		logs.Info(err, "create it")
		outFile, err = os.Create(*logFile)
		if err != nil {
			logs.Error("open log file error %v", err)
		}
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			logs.Error("close log file err: %v", err)
		}
	}()
	cmd.Stdout = outFile
	cmd.Stderr = outFile
	if err := cmd.Start(); err != nil {
		logs.Error("fork new process err", err)
	}
	childPid = cmd.Process.Pid

	//avoid the defunct process
	go func() {
		cmd.Wait()
	}()

	logs.Info("forked pid is %s", strconv.Itoa(childPid))
}

//kill the child process
func killChildProcess() {
	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Kill(); err != nil {
			logs.Warn("kill child process err %v", err)
		}
	}
	/* option method
	if runtime.GOOS == "windows" {
		//todo need test in windows environment
		cmd := exec.Command("taskkill", "/f", "/pid", strconv.Itoa(childPid))
		if err := cmd.Run(); err != nil {
			logs.Error("kill child process failed, process pid %s", strconv.Itoa(childPid))
		}
	} else {
		// normal kill
		cmd := exec.Command("kill", strconv.Itoa(childPid))
		if err := cmd.Run(); err != nil {
			//kill with force
			cmd1 := exec.Command("kill", "-9", strconv.Itoa(childPid))
			if err := cmd1.Run(); err != nil {
				logs.Error("kill child process failed, process pid %s's", strconv.Itoa(childPid))
			}
		}
	}
	*/
}

func httpGet(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("fetch: %v", err)
		return []byte(err.Error())
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return []byte(err.Error())
	}

	return b
}
