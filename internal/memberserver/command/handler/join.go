package handler

import (
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
)

type JoinHandler struct{}

func (h *JoinHandler) Handle(args []string) (string, error) {
	instance, err := heartbeat.GetInstance()
	if err != nil {
		return "", err
	}
	if instance.IsRunning == true {
		return "Already in the group", nil
	}
	instance.Start()
	return "Start Heartbeating", nil
}
