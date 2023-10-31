package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var putCmd = &cobra.Command{
	Use:     "put [localfilename] [sdfsfilename]",
	Short:   "put file from SDFS",
	Long:    `put file from SDFS`,
	Example: `  sdfs put local_test sdfs_test`,
	Args:    cobra.ExactArgs(2),
	Run:     put,
}

func put(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	err = client.PutFile(args[0], args[1])
	if err != nil {
		logrus.Fatal(err)
	}
}

func init() {
}
