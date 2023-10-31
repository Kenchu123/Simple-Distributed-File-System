package leaderserver

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/metadata"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
)

// LeaderServer handles file operations permission and Leader election.
type LeaderServer struct {
	port          string
	leader        string
	metadata      *metadata.Metadata
	fileSemaphore map[string]*semaphore.Weighted

	pb.UnimplementedLeaderServerServer
}

// NewLeader creates a new Leader.
func NewLeaderServer(port string) *LeaderServer {
	return &LeaderServer{
		port:   port,
		leader: "localhost", // TODO: find the right leader
		metadata: &metadata.Metadata{
			FileInfo: map[string]metadata.BlockInfo{
				"test": {
					0: {
						HostNames: []string{"localhost"},
						FileName:  "test",
						BlockID:   0,
					},
					1: {
						HostNames: []string{"localhost"},
						FileName:  "test",
						BlockID:   1,
					},
				},
			},
		},
		fileSemaphore: map[string]*semaphore.Weighted{
			"test": semaphore.NewWeighted(2),
		},
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

func (l *LeaderServer) SetLeader(leader string) {
	l.leader = leader
}

// TODO: Elect Leader, Put file, get file, del file

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
