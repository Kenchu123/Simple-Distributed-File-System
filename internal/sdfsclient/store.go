package client

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Store list all files currently being stored at this machine
func (c *Client) Store() (string, error) {
	leader, err := c.getLeader()
	if err != nil {
		return "", err
	}
	logrus.Infof("Leader is %s", leader)

	metadata, err := c.getMetadata(leader)
	if err != nil {
		return "", err
	}
	logrus.Infof("metadata: %+v", metadata)

	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}
	logrus.Infof("hostName: %s", hostName)

	re := ""
	for fileName, fileInfo := range metadata.GetFileInfo() {
		for blockID, blockMeta := range fileInfo.BlockInfo {
			for _, host := range blockMeta.HostNames {
				if host == hostName {
					re += fmt.Sprintf("-- file %s, block %d\n", fileName, blockID)
				}
			}
		}
	}
	return re, nil
}
