package enable

import (
	"github.com/spf13/cobra"
)

var configPath string
var machineRegex string
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable",
	Long:  `Enable`,
}

func New() *cobra.Command {
	return enableCmd
}

func init() {
	enableCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	enableCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
	enableCmd.AddCommand(suspicionCmd)
}
