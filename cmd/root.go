package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/logger"
)

var rootCmd = &cobra.Command{
	Use:   "sdfs",
	Short: "Simple Distributed File System ",
	Long:  `Machine Programming 3 - Simple Distributed File System `,
}
var logPath string
var configPath string

func Execute() error {
	logger.Init(logPath)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logPath, "log", "l", "logs/sdfs.log", "path to log file")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")

	rootCmd.AddCommand(serveCmd, getCmd)
}
