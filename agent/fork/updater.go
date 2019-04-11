package fork

import (
	"encoding/json"
	"flag"
	"github.com/waves-zhangyt/kiteagent/agent/conf"
	"github.com/waves-zhangyt/kiteagent/agent/util/logs"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var updater = flag.Bool("updater", false, "update app available when start with -daemon=true")

var updating bool
var updatingMu sync.Mutex

type UpdateConfig struct {
	Enabled               bool   `json:"enabled", yaml:"enabled"`
	CheckIntervalSeconds  int    `json:"checkIntervalSeconds" yaml:"checkIntervalSeconds"`
	LatestAgentVersionUrl string `json:"latestAgentVersionUrl" yaml:"latestAgentVersionUrl"`
	LatestAgentUrl        string `json:"latestAgentUrl" yaml:"latestAgentUrl"`
	UpdateReloadPort      int    `json:"updateReloadPort" yaml:"updateReloadPort"`
}

var DefaultUpdateConfig UpdateConfig

func InitConfig() {
	data, err := ioutil.ReadFile("conf/kite-agent-updater.yml")
	if err != nil {
		logs.Error("read kite-agent-updater.yml err: %v", err)
	}
	if err := yaml.Unmarshal(data, &DefaultUpdateConfig); err != nil {
		logs.Error("yaml unmarshaling failed: %v", err)
	}
}

// check need update and execute update
func checkUpdate() {
	for {
		interval := time.Duration(DefaultUpdateConfig.CheckIntervalSeconds)
		time.Sleep(interval * time.Second)

		if DefaultUpdateConfig.Enabled {
			conf.SyncLoadConfig()

			// current version, version pattern : v0.7.0 is's setting in httpserver/httpserver.go
			version := httpGet("http://127.0.0.1:" + strconv.Itoa(conf.DefaultConfig.HttpServerPort) + "/version")
			//kite manager version
			kiteManagerVersion := httpGet(DefaultUpdateConfig.LatestAgentVersionUrl)
			//when different with kite manager version, use the kitemanager version, that means version can rollback
			if needUpdate(string(version), string(kiteManagerVersion)) {
				logs.Debug("do update to version %v", string(kiteManagerVersion))
				update()
			}
		}
	}
}

func needUpdate(version, managerVersion string) bool {
	//when can't get the version info or not a version info from the endpoint
	if !strings.HasPrefix(managerVersion, "v") || !strings.HasPrefix(version, "v") {
		return false
	}

	return version != managerVersion
}

func IsUpdating() bool {
	updatingMu.Lock()
	defer updatingMu.Unlock()
	return updating
}

func setUpdating(v bool) {
	updatingMu.Lock()
	defer updatingMu.Unlock()
	updating = v
}

func update() {
	setUpdating(true)
	defer setUpdating(false)

	// get the agent app file path
	file, _ := exec.LookPath(os.Args[0])
	agentPath, _ := filepath.Abs(file)

	logs.Info("remove the old version")
	if err := cmd.Process.Kill(); err != nil {
		logs.Error("kill current version process err: %v", err)
		// try force kill when err account
		if err3 := exec.Command("bash", "-c", "kill -9 "+strconv.Itoa(cmd.Process.Pid)).Run(); err3 != nil {
			logs.Error("force kill for update err: %v", err3)
		}
		return
	}
	if err1 := os.Remove(agentPath); err1 != nil {
		logs.Error("reomve old version app file err: %v", err1)
	}
	logs.Info("remove the old version done")

	logs.Info("get new version")
	agentBytes := httpGet(DefaultUpdateConfig.LatestAgentUrl)
	if err2 := ioutil.WriteFile(agentPath, agentBytes, 0766); err2 != nil {
		logs.Error("write new version app err: %v", err2)
	}
	logs.Info("get new version done")

	// let the daemon fork to start the new version
}

func init() {

	if !flag.Parsed() {
		flag.Parse()
	}

	// must start with -daemon=true and -background=false
	if *updater && *daemon && !*background {
		InitConfig()

		// enable config reload endpoint
		go func() {
			http.HandleFunc("/reloadConfig", func(writer http.ResponseWriter, request *http.Request) {
				InitConfig()
				writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
				io.WriteString(writer, "ok")
			})
			http.HandleFunc("/config", func(writer http.ResponseWriter, request *http.Request) {
				InitConfig()
				writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
				b, _ := json.Marshal(DefaultUpdateConfig)
				io.WriteString(writer, string(b))
			})
			// 注意这里把端口写固定了，以后可能会写成可配置的
			log.Fatal(http.ListenAndServe("localhost:19989", nil))
		}()

		go checkUpdate()
	}

}
