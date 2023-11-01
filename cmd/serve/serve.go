package serve

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/sdfsserver"
)

var configPath string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Long:  `Start the server`,
	Run:   serve,
}

func serve(cmd *cobra.Command, args []string) {
	server, err := sdfsserver.NewServer(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	server.Run()
}

func New() *cobra.Command {
	return serveCmd
}

func init() {
	serveCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
}
