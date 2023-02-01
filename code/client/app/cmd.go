package app

import (
	"fmt"
	"os"
	"path/filepath"
	rt "runtime"

	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule/vnc"
	"github.com/lwch/natpass/code/utils"
	"github.com/lwch/runtime"
	"github.com/spf13/cobra"
)

// User --user param
var User string

// ConfDir --conf param
var ConfDir string

// Version application version
var Version string

// vnc child process params
var (
	VncName   string
	VncPort   uint16
	VncCursor bool
)

// App application
type App struct {
	p *program
}

// NewApp create application
func NewApp() *App {
	return &App{
		p: newProgram(),
	}
}

func buildService(p *program) service.Service {
	dir, err := filepath.Abs(ConfDir)
	runtime.Assert(err)

	var depends []string
	if rt.GOOS != "windows" {
		depends = append(depends, "After=network.target")
	}
	var opt service.KeyValue
	switch rt.GOOS {
	case "windows":
		opt = service.KeyValue{
			"StartType":              "automatic",
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "5s",
			"OnFailureResetPeriod":   10,
		}
	case "linux":
		opt = service.KeyValue{
			"LimitNOFILE": 65000,
		}
	case "darwin":
		opt = service.KeyValue{
			"SessionCreate": true,
		}
	}

	svc, err := service.New(p, &service.Config{
		Name:         "np-cli",
		DisplayName:  "np-cli",
		Description:  "natpass client",
		UserName:     User,
		Arguments:    []string{"--conf", dir},
		Dependencies: depends,
		Option:       opt,
	})
	runtime.Assert(err)
	return svc
}

// Run run application
func (a *App) Run(*cobra.Command, []string) {
	a.p.
		setConfDir(ConfDir).
		setConfigure(global.LoadConf(ConfDir))

	runtime.Assert(buildService(a.p).Run())
}

// Install register service
func (a *App) Install(*cobra.Command, []string) {
	cfg := global.LoadConf(ConfDir)

	err := buildService(a.p).Install()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	utils.BuildDir(cfg.LogDir, User)
	utils.BuildDir(cfg.CodeDir, User)
	fmt.Println("register service np-cli success")
}

// Uninstall unregister service
func (a *App) Uninstall(*cobra.Command, []string) {
	err := buildService(a.p).Uninstall()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("unregister service np-cli success")
}

// Start start service
func (a *App) Start(*cobra.Command, []string) {
	err := buildService(a.p).Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("start service np-cli success")
}

// Stop stop service
func (a *App) Stop(*cobra.Command, []string) {
	err := buildService(a.p).Stop()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("stop service np-cli success")
}

// Restart restart service
func (a *App) Restart(*cobra.Command, []string) {
	err := buildService(a.p).Restart()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("restart service np-cli success")
}

// Status show service status
func (a *App) Status(*cobra.Command, []string) {
	status, err := buildService(a.p).Status()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	switch status {
	case service.StatusRunning:
		fmt.Println("service is running")
	case service.StatusStopped:
		fmt.Println("service is stopped")
	case service.StatusUnknown:
		fmt.Println("service status is unknown")
	}
}

// Vnc handle vnc child process handler
func (a *App) Vnc(*cobra.Command, []string) {
	defer utils.Recover("vnc.worker")

	cfg := global.LoadConf(ConfDir)

	stdout := true
	if rt.GOOS == "windows" {
		stdout = false
	}
	// go func() {
	// 	http.ListenAndServe(":9001", nil)
	// }()
	logging.SetSizeRotate(logging.SizeRotateConfig{
		Dir:         cfg.LogDir,
		Name:        "np-cli.vnc." + VncName,
		Size:        int64(cfg.LogSize.Bytes()),
		Rotate:      cfg.LogRotate,
		WriteStdout: stdout,
		WriteFile:   true,
	})
	defer logging.Flush()
	vnc.RunWorker(VncPort, VncCursor)
}
