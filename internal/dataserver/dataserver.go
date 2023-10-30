package dataserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/metadata"
	"google.golang.org/grpc"
)

// DataServer handle data blocks and metadata.
type DataServer struct {
	port      string
	blocksDir string
	metaData  *metadata.Metadata

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
		metaData:  metadata.NewMetadata(),
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

// GetFileBlocks gets all file blocks of a file and returns them to the client through gRPC.
func (ds *DataServer) GetFileBlocks(ctx context.Context, in *pb.GetFileBlocksRequest) (*pb.GetFileBlocksReply, error) {
	fileName := in.GetFileName()
	dataBlocks, err := ds.getFileBlocks(fileName)
	if err != nil {
		return nil, err
	}
	reply := &pb.GetFileBlocksReply{
		FileName:   fileName,
		DataBlocks: []*pb.GetFileBlocksReply_DataBlocks{},
	}
	for _, dataBlock := range dataBlocks.dataBlocks {
		reply.DataBlocks = append(reply.DataBlocks, &pb.GetFileBlocksReply_DataBlocks{
			BlockID: dataBlock.BlockID,
			Data:    dataBlock.Data,
		})
	}
	return reply, nil
}

// getFileBlocks gets all file blocks of a file.
func (ds *DataServer) getFileBlocks(filename string) (*DataBlocks, error) {
	// get fileBlocks from metadata using filename
	// for each fileBlock, read data from filepath

	// TODO: DELETE THIS PART
	ds.metaData.FileInfo["test"] = map[int64]metadata.BlockMeta{
		0: {
			FileName:  "test",
			BlockID:   0,
			BlockSize: 10,
		},
		1: {
			FileName:  "test",
			BlockID:   1,
			BlockSize: 15,
		},
	}
	blockInfo, err := ds.metaData.GetBlockInfo(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to getFileBlocks: %v", err)
	}
	dataBlocks := &DataBlocks{fileName: filename}
	for blockID, blockMeta := range blockInfo {
		filePath := filepath.Join(ds.blocksDir, filename+"_"+strconv.Itoa(int(blockID)))
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
		}
		dataBlocks.dataBlocks = append(dataBlocks.dataBlocks, DataBlock{blockMeta.BlockID, data})
	}
	return dataBlocks, nil
}

// GetFileBlock gets a file block of a file and returns it to the client through gRPC.
func (ds *DataServer) GetFileBlock(ctx context.Context, in *pb.GetFileBlockRequest) (*pb.GetFileBlockReply, error) {
	// TODO: DELETE THIS PART
	ds.metaData.FileInfo["test"] = map[int64]metadata.BlockMeta{
		0: {
			FileName: "test",
			BlockID:  0,
		},
		1: {
			FileName: "test",
			BlockID:  1,
		},
	}
	fileName := in.GetFileName()
	blockID := in.GetBlockID()
	dataBlock, err := ds.getFileBlock(fileName, blockID)
	if err != nil {
		return nil, err
	}
	reply := &pb.GetFileBlockReply{
		Data: dataBlock.Data,
	}
	return reply, nil
}

func (ds *DataServer) getFileBlock(fileName string, blockID int64) (*DataBlock, error) {
	// get fileBlock from metadata using filename and blockID
	// read data from filepath
	// return dataBlock
	blockMeta, err := ds.metaData.GetBlockMeta(fileName, blockID)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(ds.blocksDir, blockMeta.FileName+"_"+strconv.Itoa(int(blockMeta.BlockID)))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return &DataBlock{blockMeta.BlockID, data}, nil
}

func (ds *DataServer) PutFileBlock(ctx context.Context, in *pb.PutFileBlockRequest) (*pb.PutFileBlockReply, error) {
	fileName := in.GetFileName()
	blockID := in.GetBlockID()
	blockSize := in.GetBlockSize()
	data := in.GetData()
	err := ds.putFileBlock(fileName, int(blockID), int(blockSize), data)
	if err != nil {
		return nil, err
	}
	reply := &pb.PutFileBlockReply{
		Ok: true,
	}
	return reply, nil
}

func (ds *DataServer) putFileBlock(fileName string, blockID int, blockSize int, data []byte) error {
	filePath := filepath.Join(ds.blocksDir, fileName, "_", strconv.Itoa(blockID))
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	// TODO: commit to update metadata
	return nil
}
