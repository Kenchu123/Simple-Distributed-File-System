package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DelFile deletes a file from SDFS.
func (c *Client) DelFile(sdfsfilename string) error {
	leader, err := c.getLeader()
	if err != nil {
		return err
	}
	logrus.Infof("Leader is %s", leader)

	// acquire write lock
	err = c.acquireFileWriteLock(leader, sdfsfilename)
	if err != nil {
		return err
	}
	defer c.releaseFileWriteLock(leader, sdfsfilename)
	logrus.Infof("Acquired write lock of file %s", sdfsfilename)

	err = c.delFile(leader, sdfsfilename)
	if err != nil {
		return err
	}
	logrus.Infof("Deleted file %s from SDFS", sdfsfilename)
	return nil
}

// delFile from local leader server through gRPC.
func (c *Client) delFile(leader, sdfsfilename string) error {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot dial leader server %s: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.DelFile(ctx, &leaderServerProto.DelFileRequest{
		FileName: sdfsfilename,
	})
	if err != nil {
		return fmt.Errorf("cannot delete file %s: %v", sdfsfilename, err)
	}
	return nil
}
