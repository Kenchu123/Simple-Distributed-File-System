package dataserver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

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
	// clean up blocksDir
	if err := filepath.Walk(blocksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("Failed to walk blocksDir %s: %v", blocksDir, err)
			return err
		}
		if !info.IsDir() {
			if err := os.Remove(path); err != nil {
				logrus.Errorf("Failed to remove file %s: %v", path, err)
				return err
			}
		}
		return nil
	}); err != nil {
		logrus.Errorf("Failed to clean up blocksDir %s: %v", blocksDir, err)
		return nil
	}

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

func (ds *DataServer) GetFilePath(fileName string, blockID int64) string {
	return filepath.Join(ds.blocksDir, fmt.Sprintf("%s_%d", fileName, blockID))
}
