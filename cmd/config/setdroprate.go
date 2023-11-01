package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/client"
)

var dropRate float32
var dropRateCmd = &cobra.Command{
	Use:   "set-droprate",
	Short: "Manage droprate",
	Long:  "Manage droprate",
	Run:   setDroprate,
}

func setDroprate(cmd *cobra.Command, args []string) {
	client, err := client.New(configPath, machineRegex)
	if err != nil {
		logrus.Fatalf("failed to create command client: %v", err)
	}
	results := client.Run([]string{string(command.DROPRATE), fmt.Sprintf("%f", dropRate)})
	for _, r := range results {
		if r.Err != nil {
			logrus.Errorf("failed to send command to %s: %v\n", r.Hostname, r.Err)
			continue
		}
		logrus.Printf("%s: %s\n", r.Hostname, r.Message)
	}
}

func init() {
	dropRateCmd.Flags().Float32VarP(&dropRate, "droprate", "d", 0.0, "droprate")
}
