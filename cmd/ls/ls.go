package ls

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var configPath string

var lsCmd = &cobra.Command{
	Use:     "ls [sdfsfilename]",
	Short:   "list all machine (VM) addresses where this file is currently being stored",
	Long:    `list all machine (VM) addresses where this file is currently being stored`,
	Example: `  sdfs ls sdfs_test`,
	Args:    cobra.ExactArgs(1),
	Run:     ls,
}

func ls(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	re, err := client.LsFile(args[0])
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf("file %s's block location:\n%s\n", args[0], re)
}

func New() *cobra.Command {
	return lsCmd
}

func init() {
	lsCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
}
