package handler

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type VerboseHandler struct{}

func (h *VerboseHandler) Handle(args []string) (string, error) {
	verbose := args[1] == "true"
	if verbose {
		logrus.Infof("[LOG] Set log level to DebugLevel")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.Infof("[LOG] Set log level to InfoLevel")
		logrus.SetLevel(logrus.InfoLevel)
	}
	return fmt.Sprintf("Set log level to %t", verbose), nil
}
