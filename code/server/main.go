package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/lwch/natpass/code/server/app"
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

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringVarP(&app.ConfDir, "conf", "c", "", "configure file path")
	rootCmd.MarkFlagRequired("conf")
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
