package dataserver

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
)

func (ds *DataServer) GetFileBlock(in *pb.GetFileBlockRequest, stream pb.DataServer_GetFileBlockServer) error {
	fileName := in.GetFileName()
	blockID := in.GetBlockID()
	file, err := ds.openFile(fileName, blockID)
	if err != nil {
		return err
	}
	buf := make([]byte, CHUNK_SIZE)
	fileSize := 0
	for {
		num, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := buf[:num]
		if err := stream.Send(&pb.GetFileBlockReply{Chunk: chunk}); err != nil {
			return err
		}
		fileSize += len(chunk)
		logrus.Debugf("sent a chunk with size %v", len(chunk))
	}
	logrus.Infof("sent file %s block %d with size %d", fileName, blockID, fileSize)
	return nil
}

func (ds *DataServer) openFile(fileName string, blockID int64) (*os.File, error) {
	// get fileBlock from metadata using filename and blockID
	// read data from filepath
	// return dataBlock
	filePath := ds.GetFilePath(fileName, blockID)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	return file, nil
}
