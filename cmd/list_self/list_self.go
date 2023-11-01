package list_self

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/client"
)

var configPath string
var machineRegex string

var selfCmd = &cobra.Command{
	Use:   "list_self",
	Short: "Show the id of the current node",
	Long:  `Show the id of the current node`,
	Run:   list_self,
}

func list_self(cmd *cobra.Command, args []string) {
	client, err := client.New(configPath, machineRegex)
	if err != nil {
		logrus.Fatalf("failed to create command client: %v", err)
	}
	results := client.Run([]string{string(command.ID)})
	for _, r := range results {
		if r.Err != nil {
			logrus.Errorf("failed to send command to %s: %v\n", r.Hostname, r.Err)
			continue
		}
		logrus.Printf("%s: %s\n", r.Hostname, r.Message)
	}
}

func New() *cobra.Command {
	return selfCmd
}

func init() {
	selfCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	selfCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
}
