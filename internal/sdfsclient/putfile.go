package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	dataServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var CHUNK_SIZE = 3 * 1024 * 1024

// PutFile sends a file to the SDFS.
func (c *Client) PutFile(localfilename, sdfsfilename string) error {
	localfile, err := os.Open(localfilename)
	if err != nil {
		return fmt.Errorf("cannot open local file %s: %v", localfilename, err)
	}
	defer localfile.Close()
	fileInfo, err := localfile.Stat()
	if err != nil {
		return fmt.Errorf("cannot get local file %s info: %v", localfilename, err)
	}

	// get leader, ask leader where to store the file, send the file to the data server
	leader, err := c.getLeader()
	if err != nil {
		return err
	}
	logrus.Infof("Leader is %s", leader)

	blockInfo, err := c.putBlockInfo(leader, sdfsfilename, fileInfo.Size())
	defer c.putFileOK(leader, sdfsfilename, blockInfo)
	if err != nil {
		return err
	}
	logrus.Infof("Got blockInfo %+v", blockInfo)

	eg, _ := errgroup.WithContext(context.Background())
	for i := int64(0); i < int64(len(blockInfo)); i++ {
		// read a block from localfile
		block := make([]byte, c.blockSize)
		n, err := localfile.Read(block)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read local file %s: %v", localfilename, err)
		}
		logrus.Infof("Read block %d of file %s with size %d", i, localfilename, n)
		for _, hostname := range blockInfo[i].HostNames {
			// send the block to the data server
			func(hostname, sdfsfilename string, blockID int64, block []byte) {
				eg.Go(func() error {
					_, err := c.putFileBlock(hostname, sdfsfilename, blockID, block)
					if err != nil {
						return fmt.Errorf("Failed to put block %d of file %s to data server %s with error %w", blockID, sdfsfilename, hostname, err)
					}
					logrus.Infof("Put block %d of file %s to data server %s", blockID, sdfsfilename, hostname)
					return nil
				})
			}(hostname, sdfsfilename, i, block[:n])
		}
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("Failed to put file %s to SDFS with error %w", sdfsfilename, err)
	}
	logrus.Infof("Put all blocks of file %s to SDFS", sdfsfilename)
	// TODO: putblockOK
	return nil
}

// putBlockInfo gets the block info for putting a file from the leader server.
func (c *Client) putBlockInfo(leader, fileName string, fileSize int64) (metadata.BlockInfo, error) {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.PutBlockInfo(ctx, &leaderServerProto.PutBlockInfoRequest{
		FileName: fileName,
		FileSize: fileSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get block info: %v", err)
	}
	putBlockInforReplyBlockMeta := r.GetBlockInfo()
	blockInfo := metadata.BlockInfo{}
	for blockID, blockMeta := range putBlockInforReplyBlockMeta {
		blockInfo[blockID] = metadata.BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	return blockInfo, nil
}

// putFileBlock sends the file block to the data server.
func (c *Client) putFileBlock(hostname, fileName string, blockID int64, data []byte) (bool, error) {
	conn, err := grpc.Dial(hostname+":"+c.dataServerPort, []grpc.DialOption{
		grpc.WithInitialWindowSize(1024 * 1024 * 1024),
		grpc.WithInitialConnWindowSize(1024 * 1024 * 1024),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)

	if err != nil {
		return false, fmt.Errorf("cannot connect to dataServer: %v", err)
	}
	defer conn.Close()

	client := dataServerProto.NewDataServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	stream, err := client.PutFileBlock(ctx)
	if err != nil {
		return false, err
	}
	var fileSize int64 = 0
	for {
		if len(data) == 0 {
			break
		}
		chunk := []byte(data)
		if len(chunk) > CHUNK_SIZE {
			chunk = chunk[:CHUNK_SIZE]
		}
		if err := stream.Send(&dataServerProto.PutFileBlockRequest{
			FileName: fileName,
			BlockID:  blockID,
			Chunk:    chunk,
		}); err != nil {
			return false, err
		}
		fileSize += int64(len(chunk))
		data = data[len(chunk):]
	}
	r, err := stream.CloseAndRecv()
	logrus.Debugf("sent file %s block %d with size %d", fileName, blockID, fileSize)
	return r.GetOk(), err
}

// putFileOK tells the leader server that the client has put the file.
func (c *Client) putFileOK(hostname, fileName string, blockInfo metadata.BlockInfo) error {
	conn, err := grpc.Dial(hostname+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", hostname, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	putFileOKRequestBlockInfo := map[int64]*leaderServerProto.BlockMeta{}
	for blockID, blockMeta := range blockInfo {
		putFileOKRequestBlockInfo[blockID] = &leaderServerProto.BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	_, err = client.PutFileOK(ctx, &leaderServerProto.PutFileOKRequest{
		FileName:  fileName,
		BlockInfo: putFileOKRequestBlockInfo,
	})
	if err != nil {
		return fmt.Errorf("failed to put file ok of %s: %v", fileName, err)
	}
	return nil
}
