package client

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LsFile list all machine (VM) addresses where this file is currently being stored
func (c *Client) LsFile(sdfsfilename string) (string, error) {
	leader, err := c.getLeader()
	if err != nil {
		return "", err
	}
	logrus.Infof("Leader is %s", leader)

	metadata, err := c.getMetadata(leader)
	if err != nil {
		return "", err
	}

	blockInfo, err := metadata.GetBlockInfo(sdfsfilename)
	if err != nil {
		return "", err
	}
	re := ""
	for _, blockMeta := range blockInfo {
		re += fmt.Sprintf("-- block %d: ", blockMeta.BlockID)
		for _, hostName := range blockMeta.HostNames {
			re += fmt.Sprintf("%s ", hostName)
		}
		re += "\n"
	}

	return re, nil
}
