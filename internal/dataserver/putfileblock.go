package dataserver

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
)

func (ds *DataServer) PutFileBlock(stream pb.DataServer_PutFileBlockServer) error {
	var fileName string
	var blockID int64
	buffer := make([]byte, 0)
	var fileSize int64 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk from server: %v", err)
		}
		fileName = req.GetFileName()
		blockID = req.GetBlockID()
		chunk := req.GetChunk()
		fileSize += int64(len(chunk))
		logrus.Debugf("received a chunk with size %v", len(chunk))
		buffer = append(buffer, chunk...)
	}
	err := ds.writeFileBlock(fileName, blockID, buffer)
	if err != nil {
		return err
	}
	logrus.Infof("received file %s block %d with size %d", fileName, blockID, fileSize)
	return stream.SendAndClose(&pb.PutFileBlockReply{Ok: true})
}

func (ds *DataServer) writeFileBlock(fileName string, blockID int64, data []byte) error {
	filePath := ds.GetFilePath(fileName, blockID)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	return nil
}
