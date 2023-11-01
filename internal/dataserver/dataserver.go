package dataserver

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"google.golang.org/grpc"
)

var CHUNK_SIZE = 3 * 1024 * 1024

// DataServer handle data blocks and metadata.
type DataServer struct {
	port      string
	blocksDir string

	pb.UnimplementedDataServerServer
}

type DataBlocks struct {
	fileName   string
	dataBlocks []DataBlock
}

type DataBlock struct {
	BlockID int64
	Data    []byte
}

// NewDataServer creates a new dataserver.
func NewDataServer(port, blocksDir string) *DataServer {
	return &DataServer{
		blocksDir: blocksDir,
		port:      port,
	}
}

// RunDataServer run the dataserver
func (ds *DataServer) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", ds.port))
	if err != nil {
		logrus.Fatalf("failed to listen on port %s: %v\n", ds.port, err)
		return
	}
	defer listen.Close()
	grpcServer := grpc.NewServer()
	pb.RegisterDataServerServer(grpcServer, ds)
	logrus.Infof("DataServer listening on port %s", ds.port)
	if err := grpcServer.Serve(listen); err != nil {
		logrus.Fatalf("failed to serve: %v\n", err)
		return
	}
}

func (ds *DataServer) GetFileBlock(in *pb.GetFileBlockRequest, stream pb.DataServer_GetFileBlockServer) error {
	fileName := in.GetFileName()
	blockID := in.GetBlockID()
	file, err := ds.getFileBlock(fileName, blockID)
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
	logrus.Debugf("sent file %s block %d with size %d", fileName, blockID, fileSize)
	return nil
}

func (ds *DataServer) getFileBlock(fileName string, blockID int64) (*os.File, error) {
	// get fileBlock from metadata using filename and blockID
	// read data from filepath
	// return dataBlock
	filePath := filepath.Join(ds.blocksDir, fileName+"_"+strconv.Itoa(int(blockID)))
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	return file, nil
}

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
	logrus.Debugf("received file %s block %d with size %d", fileName, blockID, fileSize)
	err := ds.putFileBlock(fileName, blockID, buffer)
	if err != nil {
		return err
	}
	return stream.SendAndClose(&pb.PutFileBlockReply{Ok: true})
}

func (ds *DataServer) putFileBlock(fileName string, blockID int64, data []byte) error {
	filePath := filepath.Join(ds.blocksDir, fileName+"_"+strconv.Itoa(int(blockID)))
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	return nil
}
