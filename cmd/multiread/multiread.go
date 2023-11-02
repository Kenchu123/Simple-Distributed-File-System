package multiread

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/command"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
)

var configPath string
var machineRegex string
var multireadCmd = &cobra.Command{
	Use:     "multiread [sdfsfilename] [localfilename] [flags]",
	Short:   "multiread a file from SDFS",
	Long:    `multiread launches multiple machines to read a file from SDFS`,
	Example: `  sdfs multiread sdfs_test local_test -m "0[1-9]"`,
	Args:    cobra.ExactArgs(2),
	Run:     multiread,
}

func multiread(cmd *cobra.Command, args []string) {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	machines, err := conf.FilterMachines(machineRegex)
	if err != nil {
		logrus.Fatal(err)
	}
	// launch multiple machines to call get
	var wg = &sync.WaitGroup{}
	for _, machine := range machines {
		wg.Add(1)
		go func(hostname string, port string) {
			defer wg.Done()
			logrus.Infof("Launching machine %s", hostname)
			client := command.NewCommandClient(hostname, port)
			_, err := client.ExecuteCommand("get", args)
			if err != nil {
				logrus.Errorf("Error in machine %s: %v", hostname, err)
			}
			logrus.Infof("Finished get at machine %s", hostname)
		}(machine.Hostname, conf.CommandServerPort)
	}
	wg.Wait()
}

func New() *cobra.Command {
	return multireadCmd
}

func init() {
	multireadCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".sdfs/config.yml", "path to config file")
	multireadCmd.PersistentFlags().StringVarP(&machineRegex, "machine-regex", "m", ".*", "regex for machines to join (e.g. \"0[1-9]\")")
}
