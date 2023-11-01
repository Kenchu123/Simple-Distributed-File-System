package handler

import (
	"fmt"
	"strconv"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
)

type DropRateHandler struct{}

func (h *DropRateHandler) Handle(args []string) (string, error) {
	instance, err := heartbeat.GetInstance()
	if err != nil {
		return "", err
	}
	dropRate, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		return "", err
	}
	instance.SetDropRate(float32(dropRate))
	return fmt.Sprintf("Set dropRate to %f", dropRate), nil
}
