package config

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/client"
)

var verbose bool
var verboseCmd = &cobra.Command{
	Use:   "set-verbose",
	Short: "Set verbose",
	Long:  `Set verbose`,
	Run:   setVerbose,
}

func setVerbose(cmd *cobra.Command, args []string) {
	client, err := client.New(configPath, machineRegex)
	if err != nil {
		logrus.Fatalf("failed to create command client: %v", err)
	}
	results := client.Run([]string{string(command.VERBOSE), strconv.FormatBool(verbose)})
	for _, r := range results {
		if r.Err != nil {
			logrus.Errorf("failed to send command to %s: %v\n", r.Hostname, r.Err)
			continue
		}
		logrus.Printf("%s: %s\n", r.Hostname, r.Message)
	}
}

func init() {
	verboseCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable or disable verbose")
}
