package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var getCmd = &cobra.Command{
	Use:     "get [sdfsfilename] [localfilename]",
	Short:   "get file from SDFS",
	Long:    `get file from SDFS`,
	Example: `  sdfs get sdfs_test local_test`,
	Args:    cobra.ExactArgs(2),
	Run:     get,
}

func get(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	err = client.GetFile(args[0], args[1])
	if err != nil {
		logrus.Fatal(err)
	}
}

func init() {
}
