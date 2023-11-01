package handler

import (
	"os"
)

type FailHandler struct{}

func (h *FailHandler) Handle(args []string) (string, error) {
	os.Exit(1)
	return "Failing", nil
}
