package client

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (c *Client) LsFile(sdfsfilename string) (string, error) {
	// get leader
	leader, err := c.getLeader()
	if err != nil {
		return "", err
	}
	logrus.Infof("Leader is %s", leader)

	metadata, err := c.getMetadata(leader)
	if err != nil {
		return "", err
	}

	if _, ok := metadata.FileInfo[sdfsfilename]; !ok {
		return "", fmt.Errorf("file %s not found", sdfsfilename)
	}

	re := ""
	for _, blockMeta := range metadata.FileInfo[sdfsfilename] {
		re += fmt.Sprintf("-- block %d: ", blockMeta.BlockID)
		for _, hostName := range blockMeta.HostNames {
			re += fmt.Sprintf("%s ", hostName)
		}
		re += "\n"
	}

	return re, nil
}
