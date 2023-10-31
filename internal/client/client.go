package client

import (
	"context"
	"fmt"
	"time"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
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
