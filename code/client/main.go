package main

import (
	"fmt"
	"os"

	"github.com/lwch/natpass/code/client/app"
	"github.com/spf13/cobra"
)

var (
	version      string = "0.0.0"
	gitHash      string
	gitReversion string
	buildTime    string
)

var a = app.NewApp()

var rootCmd = &cobra.Command{
	Use:   "np-cli",
	Short: "natpass client",
	Run:   a.Run,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "register service",
	Run:   a.Install,
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "unregister service",
	Run:   a.Uninstall,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start service",
	Run:   a.Start,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop service",
	Run:   a.Stop,
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart service",
	Run:   a.Restart,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show service status",
	Run:   a.Status,
}

var vncCmd = &cobra.Command{
	Use:   "vnc",
	Short: "vnc child process",
	Run:   a.Vnc,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version info",
	Run: func(*cobra.Command, []string) {
		fmt.Printf("version: v%s\ntime: %s\ncommit: %s.%s\n",
			version,
			buildTime,
			gitHash, gitReversion)
		os.Exit(0)
	},
}

func main() {
	app.Version = version
	installCmd.Flags().StringVarP(&app.ConfDir, "conf", "c", "", "configure file path")
	installCmd.Flags().StringVarP(&app.User, "user", "u", "", "service user")
	installCmd.MarkFlagRequired("conf")
	rootCmd.AddCommand(installCmd, uninstallCmd)
	rootCmd.AddCommand(startCmd, stopCmd, restartCmd, statusCmd)
	rootCmd.AddCommand(versionCmd)

	vncCmd.Flags().StringVarP(&app.ConfDir, "conf", "c", "", "configure file path")
	vncCmd.Flags().StringVar(&app.VncName, "name", "", "name for log file")
	vncCmd.Flags().Uint16Var(&app.VncPort, "port", 6155, "listen port")
	vncCmd.Flags().BoolVar(&app.VncCursor, "cursor", false, "show cursor")
	vncCmd.MarkFlagRequired("conf")
	vncCmd.MarkFlagRequired("name")
	vncCmd.Hidden = true
	rootCmd.AddCommand(vncCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringVarP(&app.ConfDir, "conf", "c", "", "configure file path")
	rootCmd.MarkFlagRequired("conf")
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// user := flag.String("user", "", "service user")
	// conf := flag.String("conf", "", "configure file path")
	// act := flag.String("action", "", "install or uninstall")
	// name := flag.String("name", "", "rule name")
	// vport := flag.Uint("vport", 6155, "vnc worker listen port")
	// vcursor := flag.Bool("vcursor", false, "vnc show cursor")
	// flag.Parse()

	// if len(*conf) == 0 {
	// 	fmt.Println("missing -conf param")
	// 	os.Exit(1)
	// }

	// // for test
	// // work := worker.NewWorker()
	// // work.TestCapture()
	// // return

	// dir, err := filepath.Abs(*conf)
	// runtime.Assert(err)

	// var depends []string
	// if rt.GOOS != "windows" {
	// 	depends = append(depends, "After=network.target")
	// }
	// var opt service.KeyValue
	// switch rt.GOOS {
	// case "windows":
	// 	opt = service.KeyValue{
	// 		"StartType":              "automatic",
	// 		"OnFailure":              "restart",
	// 		"OnFailureDelayDuration": "5s",
	// 		"OnFailureResetPeriod":   10,
	// 	}
	// case "linux":
	// 	opt = service.KeyValue{
	// 		"LimitNOFILE": 65000,
	// 	}
	// case "darwin":
	// 	opt = service.KeyValue{
	// 		"SessionCreate": true,
	// 	}
	// }

	// appCfg := &service.Config{
	// 	Name:         "np-cli",
	// 	DisplayName:  "np-cli",
	// 	Description:  "nat forward service",
	// 	UserName:     *user,
	// 	Arguments:    []string{"-conf", dir},
	// 	Dependencies: depends,
	// 	Option:       opt,
	// }

	// cfg := global.LoadConf(*conf)

	// if *act == "vnc.worker" {
	// 	defer utils.Recover("vnc.worker")
	// 	stdout := true
	// 	if rt.GOOS == "windows" {
	// 		stdout = false
	// 	}
	// 	// go func() {
	// 	// 	http.ListenAndServe(":9001", nil)
	// 	// }()
	// 	logging.SetSizeRotate(logging.SizeRotateConfig{
	// 		Dir:         cfg.LogDir,
	// 		Name:        "np-cli.vnc." + *name,
	// 		Size:        int64(cfg.LogSize.Bytes()),
	// 		Rotate:      cfg.LogRotate,
	// 		WriteStdout: stdout,
	// 		WriteFile:   true,
	// 	})
	// 	defer logging.Flush()
	// 	vnc.RunWorker(uint16(*vport), *vcursor)
	// 	return
	// }

	// app := app.New(version, *conf, cfg)
	// sv, err := service.New(app, appCfg)
	// runtime.Assert(err)

	// switch *act {
	// case "install":
	// 	runtime.Assert(sv.Install())
	// 	utils.BuildDir(cfg.LogDir, *user)
	// 	utils.BuildDir(cfg.CodeDir, *user)
	// case "uninstall":
	// 	runtime.Assert(sv.Uninstall())
	// default:
	// 	runtime.Assert(sv.Run())
	// }
}
