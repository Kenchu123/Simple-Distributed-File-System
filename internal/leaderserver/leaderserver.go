package leaderserver

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
)

// LeaderServer handles file operations permission and Leader election.
type LeaderServer struct {
	port              string
	leader            string
	metadata          *metadata.Metadata
	fileSemaphore     map[string]*semaphore.Weighted
	blockSize         int64
	replicationFactor int
	pb.UnimplementedLeaderServerServer
}

// NewLeader creates a new Leader.
func NewLeaderServer(port string, blockSize int64, replicationFactor int) *LeaderServer {
	return &LeaderServer{
		port:              port,
		leader:            "fa23-cs425-8701.cs.illinois.edu", // TODO: find the right leader
		metadata:          metadata.NewMetadata(),
		fileSemaphore:     map[string]*semaphore.Weighted{},
		blockSize:         blockSize,
		replicationFactor: replicationFactor,
	}
}

// Run starts the Leader.
func (l *LeaderServer) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", l.port))
	if err != nil {
		logrus.Fatalf("failed to listen on port %s: %v\n", l.port, err)
		return
	}
	defer listen.Close()
	grpcServer := grpc.NewServer()
	pb.RegisterLeaderServerServer(grpcServer, l)
	logrus.Infof("LeaderServer listening on port %s", l.port)
	if err := grpcServer.Serve(listen); err != nil {
		logrus.Fatalf("failed to serve: %v\n", err)
		return
	}
}

// GetLeader returns the leader to the client through gRPC.
func (l *LeaderServer) GetLeader(ctx context.Context, in *pb.GetLeaderRequest) (*pb.GetLeaderReply, error) {
	leader, err := l.getLeader()
	if err != nil {
		return nil, err
	}
	return &pb.GetLeaderReply{Leader: leader}, nil
}

// getLeader returns the leader.
func (l *LeaderServer) getLeader() (string, error) {
	return l.leader, nil
}

// GetMetadata returns the metadata to the client through gRPC.
func (l *LeaderServer) GetMetadata(ctx context.Context, in *pb.GetMetadataRequest) (*pb.GetMetadataReply, error) {
	metadata := l.getMetadata()
	getMetadaReply := &pb.GetMetadataReply{
		FileInfo: map[string]*pb.BlockInfo{},
	}
	for fileName, blockInfo := range metadata.FileInfo {
		getMetadaReply.FileInfo[fileName] = &pb.BlockInfo{
			BlockInfo: map[int64]*pb.BlockMeta{},
		}
		for blockID, blockMeta := range blockInfo {
			getMetadaReply.FileInfo[fileName].BlockInfo[blockID] = &pb.BlockMeta{
				HostNames: blockMeta.HostNames,
				FileName:  blockMeta.FileName,
				BlockID:   blockMeta.BlockID,
			}
		}
	}
	return getMetadaReply, nil
}

// getMetadata returns the metadata.
func (l *LeaderServer) getMetadata() *metadata.Metadata {
	return l.metadata
}

func (l *LeaderServer) SetLeader(leader string) {
	l.leader = leader
}

// TODO: Elect Leader

func (l *LeaderServer) acquireFileSemaphore(fileName string, weight int64) error {
	if _, ok := l.fileSemaphore[fileName]; !ok {
		return fmt.Errorf("file %s not found", fileName)
	}
	l.fileSemaphore[fileName].Acquire(context.Background(), weight)
	return nil
}

func (l *LeaderServer) releaseFileSemaphore(fileName string, weight int64) {
	l.fileSemaphore[fileName].Release(weight)
}
