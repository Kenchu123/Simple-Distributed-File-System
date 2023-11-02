package client

import (
	"context"
	"fmt"
	"time"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client handles file operations to SDFS.
type Client struct {
	leaderServerPort string
	dataServerPort   string
	blockSize        int64
}

// NewClient creates a new Client.
func NewClient(configPath string) (*Client, error) {
	config, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		leaderServerPort: config.LeaderServerPort,
		dataServerPort:   config.DataServerPort,
		blockSize:        config.BlockSize,
	}, nil
}

// getLeader from local leader server through gRPC.
func (c *Client) getLeader() (string, error) {
	conn, err := grpc.Dial("localhost:"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("cannot connect to %s leaderServer: %v", "localhost", err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.GetLeader(ctx, &leaderServerProto.GetLeaderRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get leader: %v", err)
	}
	return r.GetLeader(), nil
}

func (c *Client) getMetadata(leader string) (*metadata.Metadata, error) {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.GetMetadata(ctx, &leaderServerProto.GetMetadataRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %v", err)
	}

	newMetadata := metadata.NewMetadata()
	for fileName, blockInfo := range r.GetFileInfo() {
		newBlockInfo := metadata.BlockInfo{}
		for blockID, blockMeta := range blockInfo.GetBlockInfo() {
			newBlockInfo[blockID] = metadata.BlockMeta{
				HostNames: blockMeta.HostNames,
				FileName:  blockMeta.FileName,
				BlockID:   blockMeta.BlockID,
			}
		}
		newMetadata.AddOrUpdateBlockInfo(fileName, newBlockInfo)
	}
	return newMetadata, nil
}

func (c *Client) GetMetadata() (string, error) {
	leader, err := c.getLeader()
	if err != nil {
		return "", err
	}
	metadata, err := c.getMetadata(leader)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%+v", metadata), nil
}
