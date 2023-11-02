package dataserver

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ReplicateFileBlock replicates a file block to another data server
func (ds *DataServer) ReplicateFileBlock(ctx context.Context, in *pb.ReplicateFileBlockRequest) (*pb.ReplicateFileBlockReply, error) {
	return &pb.ReplicateFileBlockReply{}, ds.replicateFileBlock(in.GetFileName(), in.GetBlockID(), in.GetTo())
}

func (ds *DataServer) replicateFileBlock(fileName string, blockID int64, to string) error {
	data, err := os.ReadFile(ds.GetFilePath(fileName, blockID))
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(to+":"+ds.port, []grpc.DialOption{
		grpc.WithInitialWindowSize(1024 * 1024 * 1024),
		grpc.WithInitialConnWindowSize(1024 * 1024 * 1024),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)

	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", to, err)
	}
	defer conn.Close()

	client := pb.NewDataServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := client.PutFileBlock(ctx)
	if err != nil {
		return err
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
		if err := stream.Send(&pb.PutFileBlockRequest{
			FileName: fileName,
			BlockID:  blockID,
			Chunk:    chunk,
		}); err != nil {
			return err
		}
		fileSize += int64(len(chunk))
		data = data[len(chunk):]
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		return err
	}
	logrus.Infof("replicated file %s block %d with size %d to %s", fileName, blockID, fileSize, to)
	return nil
}
