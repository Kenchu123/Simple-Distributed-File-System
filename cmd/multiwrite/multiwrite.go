package multiwrite

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
)

var configPath string
var machineRegex string
var multiwriteCmd = &cobra.Command{
	Use:     "multiwrite [localfilename] [sdfsfilename] [flags]",
	Short:   "multiwrite a file to SDFS",
	Long:    `multiwrite launches multiple machines to write a file to SDFS`,
	Example: `  sdfs multiwrite local_test sdfs_test -m "0[1-9]"`,
	Args:    cobra.ExactArgs(2),
	Run:     multiwrite,
}

func multiwrite(cmd *cobra.Command, args []string) {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	machines, err := conf.FilterMachines(machineRegex)
	if err != nil {
		logrus.Fatal(err)
	}
	// launch multiple machines to call write
	var wg = &sync.WaitGroup{}
	for _, machine := range machines {
		wg.Add(1)
		go func(hostname string, port string) {
			defer wg.Done()
			logrus.Infof("Launching machine %s", hostname)
			client := command.NewCommandClient(hostname, port)
			_, err := client.ExecuteCommand("put", args)
			if err != nil {
				logrus.Errorf("Error in machine %s: %v", hostname, err)
				return
			}
			logrus.Infof("Finished put at machine %s", hostname)
		}(machine.Hostname, conf.CommandServerPort)
	}
	wg.Wait()
}

func New() *cobra.Command {
	return multiwriteCmd
}

func init() {
	multiwriteCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	multiwriteCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
}
