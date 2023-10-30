package main

import (
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/cmd"
)

func main() {
	// Execute the root command
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
