package enable

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/client"
)

var suspicionCmd = &cobra.Command{
	Use:   "suspicion",
	Short: "enable suspicion",
	Long:  `enable suspicion`,
	Run:   setSuspicion,
}

func setSuspicion(cmd *cobra.Command, args []string) {
	client, err := client.New(configPath, machineRegex)
	if err != nil {
		logrus.Fatalf("failed to create command client: %v", err)
	}
	results := client.Run([]string{string(command.SUSPICION), "true"})
	for _, r := range results {
		if r.Err != nil {
			logrus.Errorf("failed to send command to %s: %v\n", r.Hostname, r.Err)
			continue
		}
		logrus.Printf("%s: %s\n", r.Hostname, r.Message)
	}
}
