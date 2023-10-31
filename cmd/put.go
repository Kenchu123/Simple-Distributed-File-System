package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put file from SDFS",
	Long:  `put file from SDFS`,
	Run:   put,
}

func put(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	err = client.PutFile("test_combine", "test")
	if err != nil {
		logrus.Fatal(err)
	}
}

func init() {
}
