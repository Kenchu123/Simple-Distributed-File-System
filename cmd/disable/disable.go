package disable

import (
	"github.com/spf13/cobra"
)

var configPath string
var machineRegex string
var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable",
	Long:  `Disable`,
}

func New() *cobra.Command {
	return disableCmd
}

func init() {
	disableCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	disableCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
	disableCmd.AddCommand(suspicionCmd)
}
