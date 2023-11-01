package config

import "github.com/spf13/cobra"

var configPath string
var machineRegex string
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage config",
	Long:  "Manage config",
}

func New() *cobra.Command {
	return configCmd
}

func init() {
	configCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	configCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
	configCmd.AddCommand(dropRateCmd, verboseCmd)
}
