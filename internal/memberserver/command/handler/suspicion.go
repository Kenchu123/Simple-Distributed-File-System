package handler

import (
	"fmt"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
)

type SuspicionHandler struct{}

func (h *SuspicionHandler) Handle(args []string) (string, error) {
	instance, err := heartbeat.GetInstance()
	if err != nil {
		return "", err
	}
	enabled := args[1] == "true"
	instance.SetSuspicion(enabled)
	return fmt.Sprintf("Set suspicion enabled to %t", enabled), nil
}
