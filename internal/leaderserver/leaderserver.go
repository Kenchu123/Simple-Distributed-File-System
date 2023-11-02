package leaderserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"google.golang.org/grpc"
)

// LeaderServer handles file operations permission and Leader election.
type LeaderServer struct {
	port           string
	dataServerPort string
	leader         string
	hostname       string
	metadata       *metadata.Metadata
	blockSize      int64

	recoverReplicaTicker     *time.Ticker
	recoverReplicaTickerDone chan bool
	replicationFactor        int

	electLeaderTicker     *time.Ticker
	electLeaderTickerDone chan bool

	syncMetadataTicker     *time.Ticker
	syncMetadataTickerDone chan bool

	pb.UnimplementedLeaderServerServer
}

// NewLeader creates a new Leader.
func NewLeaderServer(port, dataServerPort string, blockSize int64, replicationFactor int) *LeaderServer {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("failed to get hostname: %v\n", err)
		return nil
	}
	return &LeaderServer{
		port:              port,
		dataServerPort:    dataServerPort,
		leader:            "", // find the right leader
		hostname:          hostname,
		metadata:          metadata.NewMetadata(),
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
	go l.startElectingLeader()
	go l.startRecoveringReplica()
	go l.startSyncingMetadata()
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
	leader := l.getLeader()
	return &pb.GetLeaderReply{Leader: leader}, nil
}

// getLeader returns the leader.
func (l *LeaderServer) getLeader() string {
	return l.leader
}

// GetMetadata returns the metadata to the client through gRPC.
func (l *LeaderServer) GetMetadata(ctx context.Context, in *pb.GetMetadataRequest) (*pb.GetMetadataReply, error) {
	metadata := l.getMetadata()
	getMetadaReply := &pb.GetMetadataReply{
		FileInfo: map[string]*pb.BlockInfo{},
	}
	for fileName, fileInfo := range metadata.GetFileInfo() {
		getMetadaReply.FileInfo[fileName] = &pb.BlockInfo{
			BlockInfo: map[int64]*pb.BlockMeta{},
		}
		for blockID, blockMeta := range fileInfo.BlockInfo {
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

func (l *LeaderServer) SetLeader(ctx context.Context, in *pb.SetLeaderRequest) (*pb.SetLeaderReply, error) {
	l.setLeader(in.GetLeader())
	return &pb.SetLeaderReply{Ok: true}, nil
}

func (l *LeaderServer) setLeader(leader string) {
	if leader != l.leader {
		logrus.Infof("leader changed from %s to %s", l.leader, leader)
	}
	l.leader = leader
}

func (l *LeaderServer) acquireFileSemaphore(fileName string, weight int64) error {
	return l.metadata.AcquireFileSemaphore(fileName, weight)
}

func (l *LeaderServer) releaseFileSemaphore(fileName string, weight int64) {
	l.metadata.ReleaseFileSemaphore(fileName, weight)
}
