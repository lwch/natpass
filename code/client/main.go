package main

import (
	"flag"
	"fmt"
	"natpass/code/client/app"
	"natpass/code/client/global"
	"natpass/code/client/rule/vnc"
	"natpass/code/utils"
	"os"
	"path/filepath"
	rt "runtime"

	_ "net/http/pprof"

	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var (
	version      string = "0.0.0"
	gitHash      string
	gitReversion string
	buildTime    string
)

func showVersion() {
	fmt.Printf("version: v%s\ntime: %s\ncommit: %s.%s\n",
		version,
		buildTime,
		gitHash, gitReversion)
	os.Exit(0)
}

func main() {
	user := flag.String("user", "", "service user")
	conf := flag.String("conf", "", "configure file path")
	ver := flag.Bool("version", false, "show version info")
	act := flag.String("action", "", "install or uninstall")
	name := flag.String("name", "", "rule name")
	vport := flag.Uint("vport", 6155, "vnc worker listen port")
	vcursor := flag.Bool("vcursor", false, "vnc show cursor")
	flag.Parse()

	if *ver {
		showVersion()
		os.Exit(0)
	}

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	// for test
	// work := worker.NewWorker()
	// work.TestCapture()
	// return

	dir, err := filepath.Abs(*conf)
	runtime.Assert(err)

	var depends []string
	if rt.GOOS != "windows" {
		depends = append(depends, "After=network.target")
	}

	appCfg := &service.Config{
		Name:         "np-cli",
		DisplayName:  "np-cli",
		Description:  "nat forward service",
		UserName:     *user,
		Arguments:    []string{"-conf", dir},
		Dependencies: depends,
	}

	cfg := global.LoadConf(*conf)

	if *act == "vnc.worker" {
		defer utils.Recover("vnc.worker")
		stdout := true
		if rt.GOOS == "windows" {
			stdout = false
		}
		// go func() {
		// 	http.ListenAndServe(":9001", nil)
		// }()
		logging.SetSizeRotate(cfg.LogDir, "np-cli.vnc."+*name, int(cfg.LogSize.Bytes()), cfg.LogRotate, stdout)
		defer logging.Flush()
		vnc.RunWorker(uint16(*vport), *vcursor)
		return
	}

	app := app.New(version, *conf, cfg)
	sv, err := service.New(app, appCfg)
	runtime.Assert(err)

	switch *act {
	case "install":
		runtime.Assert(sv.Install())
		utils.BuildLogDir(cfg.LogDir, *user)
	case "uninstall":
		runtime.Assert(sv.Uninstall())
	default:
		runtime.Assert(sv.Run())
	}
}
