package get

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	client "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/sdfsclient"
)

var configPath string

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

func New() *cobra.Command {
	return getCmd
}

func init() {
	getCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
}
