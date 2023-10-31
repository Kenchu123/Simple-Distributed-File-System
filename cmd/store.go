package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var storeCmd = &cobra.Command{
	Use:     "store",
	Short:   "list all files currently being stored at this machine",
	Long:    `list all files currently being stored at this machine`,
	Example: `  sdfs store`,
	Run:     store,
}

func store(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	re, err := client.Store()
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf("files stored at this machine:\n%s\n", re)
}

func init() {
}
