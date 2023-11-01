package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver"
	dataServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
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
	// TODO: handle waiting timeout in the getBlockInfo
	blockInfo, err := c.getBlockInfo(leader, sdfsfilename)
	if err != nil {
		return err
	}
	logrus.Infof("Got blockInfo %+v", blockInfo)

	// get file blocks from data servers
	blocks := map[int64]dataserver.DataBlock{}
	for _, blockMeta := range blockInfo {
		// TODO: get from multiple data servers
		// TODO: error handling
		logrus.Infof("Getting block %d of file %s from data server %s", blockMeta.BlockID, blockMeta.FileName, blockMeta.HostNames[0])
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
	err = os.WriteFile(localfilename, buffers, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", localfilename, err)
	}
	logrus.Infof("Wrote file %s to local file %s", sdfsfilename, localfilename)

	err = c.getFileOK(leader, sdfsfilename)
	if err != nil {
		return err
	}
	logrus.Infof("Got file OK from leader")

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
	conn, err := grpc.Dial(hostname+":"+c.dataServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to dataServer: %v", err)
	}
	defer conn.Close()

	client := dataServerProto.NewDataServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		logrus.Infof("received a chunk with size %v", len(chunk))
		buffer = append(buffer, chunk...)
	}
	return buffer, nil
}

// getFileOK tells the leader server that the client has got the file.
func (c *Client) getFileOK(hostname, fileName string) error {
	conn, err := grpc.Dial(hostname+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", hostname, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.GetFileOK(ctx, &leaderServerProto.GetFileOKRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to get file OK: %v", err)
	}
	return nil
}