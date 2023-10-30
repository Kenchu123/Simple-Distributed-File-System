package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver"
	dataServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client handles file operations to SDFS.
type Client struct {
	leaderServerPort string
	dataServerPort   string
}

func NewClient(configPath string) (*Client, error) {
	config, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		leaderServerPort: config.LeaderServerPort,
		dataServerPort:   config.DataServerPort,
	}, nil
}

func (c *Client) GetFile(sdfsfilename, localfilename string) error {
	// get leader, ask leader where the file is stored, get the file from the data server, combine the file blocks
	// TODO: get leader

	blockInfo := metadata.BlockInfo{
		0: metadata.BlockMeta{
			HostNames: []string{"localhost"},
			FileName:  "test",
			BlockID:   0,
		},
		1: metadata.BlockMeta{
			HostNames: []string{"localhost"},
			FileName:  "test",
			BlockID:   1,
		},
	}
	logrus.Infof("Getting file %s from SDFS", sdfsfilename)
	blocks := map[int64]dataserver.DataBlock{}
	for _, blockMeta := range blockInfo {
		// TODO: get from multiple data servers
		// TODO: error handling
		data, err := c.getFileBlock(blockMeta.HostNames[0], blockMeta.FileName, blockMeta.BlockID)
		if err != nil {
			return err
		}
		logrus.Infof("Got block %d of file %s from data server %s", blockMeta.BlockID, blockMeta.FileName, blockMeta.HostNames[0])
		blocks[blockMeta.BlockID] = dataserver.DataBlock{
			BlockID: blockMeta.BlockID,
			Data:    data,
		}
	}
	logrus.Infof("Got all blocks of file %s from SDFS", sdfsfilename)

	buffers := make([]byte, 0)
	for i := 0; i < len(blockInfo); i++ {
		if _, ok := blocks[int64(i)]; !ok {
			return fmt.Errorf("block %d of file %s not found", i, sdfsfilename)
		}
		buffers = append(buffers, blocks[int64(i)].Data...)
	}
	logrus.Infof("Combined all blocks of file %s", sdfsfilename)

	logrus.Infof("Writing file %s to local file %s", sdfsfilename, localfilename)
	err := os.WriteFile(localfilename, buffers, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", localfilename, err)
	}

	return nil
}

// getFileBlock gets a block of a file from the data server.
func (c *Client) getFileBlock(hostname, filename string, blockID int64) ([]byte, error) {
	conn, err := grpc.Dial(hostname+":"+c.dataServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to dataServer: %v", err)
	}
	defer conn.Close()

	client := dataServerProto.NewDataServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.GetFileBlock(ctx, &dataServerProto.GetFileBlockRequest{FileName: filename, BlockID: blockID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file block of %s: %v", filename, err)
	}
	return r.GetData(), nil
}

func (c *Client) PutFile(localfilename, sdfsfilename string) error {
	// get leader, ask leader where to store the file, send the file to the data server
	return nil
}

func (c *Client) DeleteFile(sdfsfilename string) error {
	// get leader, ask leader where the file is stored, delete the file from the data server
	return nil
}

func (c *Client) LsFile(sdfsfilename string) error {
	// TODO: print to console
	return nil
}

func (c *Client) Store() error {
	// TODO: print to console
	return nil
}
