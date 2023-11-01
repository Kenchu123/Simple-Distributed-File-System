package delete

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var configPath string

var deleteCmd = &cobra.Command{
	Use:     "delete sdfsfilename",
	Short:   "delete a file from SDFS",
	Long:    `delete a file from SDFS`,
	Example: `  delete sdfs_test`,
	Args:    cobra.ExactArgs(1),
	Run:     delete,
}

func delete(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	err = client.DelFile(args[0])
	if err != nil {
		logrus.Fatal(err)
	}
}

func New() *cobra.Command {
	return deleteCmd
}

func init() {
	deleteCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
}
