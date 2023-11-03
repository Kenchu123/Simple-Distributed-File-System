package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	dataServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GetFile gets a file from SDFS.
func (c *Client) GetFile(sdfsfilename, localfilename string) error {
	// get leader
	leader, err := c.getLeader()
	if err != nil {
		return err
	}
	logrus.Infof("Leader is %s", leader)

	// get blockInfo from leader
	blockInfo, err := c.getBlockInfo(leader, sdfsfilename)
	if err != nil {
		return err
	}
	logrus.Infof("Got blockInfo %+v", blockInfo)

	// acquire read lock
	err = c.acquireFileReadLock(leader, sdfsfilename)
	if err != nil {
		return err
	}
	defer c.releaseFileReadLock(leader, sdfsfilename)
	logrus.Infof("Acquired read lock of file %s", sdfsfilename)

	// get file blocks from data servers
	mu := sync.Mutex{}
	getSem := semaphore.NewWeighted(10)
	tempFileName := fmt.Sprintf("%s.temp", localfilename)
	file, err := os.Create(tempFileName)
	if err != nil {
		return fmt.Errorf("failed to create temp file %s: %v", tempFileName, err)
	}
	defer os.Remove(tempFileName)
	// get the block file from multiple servers concurrently
	eg, _ := errgroup.WithContext(context.Background())
	for _, blockMeta := range blockInfo {
		func(blockMeta metadata.BlockMeta) {
			eg.Go(func() error {
				// acquire a semaphore
				err := getSem.Acquire(context.Background(), 1)
				defer getSem.Release(1)
				if err != nil {
					return err
				}
				// random order
				hostNames := blockMeta.HostNames
				rand.Shuffle(len(hostNames), func(i, j int) {
					hostNames[i], hostNames[j] = hostNames[j], hostNames[i]
				})
				// try the first hostname, and the second ..., if all fail, return error
				for _, hostName := range hostNames {
					logrus.Infof("Getting block %d of file %s from data server %s", blockMeta.BlockID, blockMeta.FileName, hostName)
					data, err := c.getFileBlock(hostName, blockMeta.FileName, blockMeta.BlockID)
					if err != nil {
						logrus.Infof("Failed to get block %d of file %s from data server %s with error %s", blockMeta.BlockID, blockMeta.FileName, hostName, err)
						continue
					}
					logrus.Infof("Got block %d of file %s from data server %s", blockMeta.BlockID, blockMeta.FileName, hostName)
					// Write the block to the local temp file
					mu.Lock()
					_, err = file.WriteAt(data, blockMeta.BlockID*c.blockSize)
					mu.Unlock()
					if err != nil {
						logrus.Infof("Failed to write block %d of file %s to local temp file %s with error %s", blockMeta.BlockID, blockMeta.FileName, tempFileName, err)
						continue
					}
					logrus.Infof("Wrote block %d of file %s from data server %s", blockMeta.BlockID, blockMeta.FileName, hostName)
					return nil
				}
				return nil
			})
		}(blockMeta)
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("Failed to get file %s from SDFS: %w", sdfsfilename, err)
	}
	logrus.Infof("Got all blocks of file %s from SDFS", sdfsfilename)
	// Move the temp file to the local file
	err = os.Rename(tempFileName, localfilename)
	if err != nil {
		return fmt.Errorf("failed to rename temp file %s to local file %s: %v", tempFileName, localfilename, err)
	}
	logrus.Infof("Got file %s from SDFS to %s", sdfsfilename, localfilename)
	return nil
}

// getBlockInfo gets the block info of a file from the leader server.
func (c *Client) getBlockInfo(leader, fileName string) (metadata.BlockInfo, error) {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.GetBlockInfo(ctx, &leaderServerProto.GetBlockInfoRequest{
		FileName: fileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get block info: %v", err)
	}
	getBlockInforReplyBlockMeta := r.GetBlockInfo()
	blockInfo := metadata.BlockInfo{}
	for blockID, blockMeta := range getBlockInforReplyBlockMeta {
		blockInfo[blockID] = metadata.BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	return blockInfo, nil
}

// getFileBlock gets a block of a file from the data server.
func (c *Client) getFileBlock(hostname, filename string, blockID int64) ([]byte, error) {
	conn, err := grpc.Dial(hostname+":"+c.dataServerPort, grpc.WithInitialWindowSize(1024*1024*1024),
		grpc.WithInitialConnWindowSize(1024*1024*1024),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to dataServer: %v", err)
	}
	defer conn.Close()

	client := dataServerProto.NewDataServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	stream, err := client.GetFileBlock(ctx, &dataServerProto.GetFileBlockRequest{FileName: filename, BlockID: blockID})
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, 0)
	var fileSize int64 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to receive chunk from server: %v", err)
		}
		chunk := req.GetChunk()
		fileSize += int64(len(chunk))
		logrus.Debugf("received a chunk with size %v", len(chunk))
		buffer = append(buffer, chunk...)
	}
	return buffer, nil
}

// getFileOK tells the leader server that the client has got the file.
func (c *Client) getFileOK(hostname, fileName string) {
	conn, err := grpc.Dial(hostname+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Errorf("failed to get file OK: %v", err)
		return
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = client.GetFileOK(ctx, &leaderServerProto.GetFileOKRequest{
		FileName: fileName,
	})
	if err != nil {
		logrus.Errorf("failed to get file OK: %v", err)
		return
	}
	return
}
