package handler

import (
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
)

type IDHandler struct{}

func (h *IDHandler) Handle(args []string) (string, error) {
	instance, err := heartbeat.GetInstance()
	if err != nil {
		return "", err
	}
	if !instance.IsRunning {
		return "Not in the membership list", nil
	}
	id := instance.GetMembership().ID
	return id, nil
}
