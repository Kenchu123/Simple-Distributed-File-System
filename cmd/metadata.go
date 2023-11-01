package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/client"
)

var metadataCmd = &cobra.Command{
	Use:     "metadata",
	Short:   "show metadata",
	Long:    `show metadata`,
	Example: `  sdfs metadata`,
	Run:     metadata,
}

func metadata(cmd *cobra.Command, args []string) {
	client, err := client.NewClient(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	re, err := client.GetMetadata()
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf("files stored at this machine:\n%s\n", re)
}

func init() {
}
