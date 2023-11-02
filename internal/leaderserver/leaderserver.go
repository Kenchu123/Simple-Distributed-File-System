package leaderserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/membership"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LeaderServer handles file operations permission and Leader election.
type LeaderServer struct {
	port              string
	leader            string
	hostname          string
	metadata          *metadata.Metadata
	fileSemaphore     map[string]*semaphore.Weighted
	blockSize         int64
	replicationFactor int

	electLeaderTicker     *time.Ticker
	electLeaderTickerDone chan bool
	pb.UnimplementedLeaderServerServer
}

// NewLeader creates a new Leader.
func NewLeaderServer(port string, blockSize int64, replicationFactor int) *LeaderServer {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("failed to get hostname: %v\n", err)
		return nil
	}
	return &LeaderServer{
		port:              port,
		leader:            "", // find the right leader
		hostname:          hostname,
		metadata:          metadata.NewMetadata(),
		fileSemaphore:     map[string]*semaphore.Weighted{},
		blockSize:         blockSize,
		replicationFactor: replicationFactor,
	}
}

// Run starts the Leader.
func (l *LeaderServer) Run() {
	go l.startElectingLeader()
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

func (l *LeaderServer) SetLeader(ctx context.Context, in *pb.SetLeaderRequest) (*pb.SetLeaderReply, error) {
	logrus.Infof("Set leader to %s", in.GetLeader())
	l.setLeader(in.GetLeader())
	return &pb.SetLeaderReply{Ok: true}, nil
}

func (l *LeaderServer) setLeader(leader string) {
	if leader != l.leader {
		logrus.Infof("leader changed from %s to %s", l.leader, leader)
	}
	l.leader = leader
}

func (l *LeaderServer) startElectingLeader() {
	logrus.Info("Start electing leader")
	l.electLeaderTicker = time.NewTicker(time.Second * 5)
	defer l.electLeaderTicker.Stop()
	for {
		select {
		case <-l.electLeaderTickerDone:
			return
		case <-l.electLeaderTicker.C:
			l.electLeader()
		}
	}
}

// Elect Leader
func (l *LeaderServer) electLeader() {
	heartbeat, err := heartbeat.GetInstance()
	if err != nil {
		logrus.Errorf("failed to get heartbeat instance: %v", err)
		return
	}
	_membership := heartbeat.GetMembership()
	if _membership == nil {
		logrus.Errorf("failed to get membership instance")
		return
	}
	members := _membership.GetAliveMembers()
	// If leader is not alive, reset the leader
	if _, ok := members[l.leader]; !ok {
		l.setLeader("")
	}
	// If there is no leader, elect a leader
	leader := l.getLeader()
	if leader == "" {
		// See if the alive node with the largest ID is me
		leader = l.electLeaderFromMembers(members)
	}
	// If I am not the leader, do nothing (wait for the leader to update me)
	if leader != l.hostname {
		return
	}
	// I am the leader, update the other nodes
	var wg sync.WaitGroup
	for _, member := range members {
		wg.Add(1)
		go func(member *membership.Member) {
			defer wg.Done()
			hostname := member.GetName()
			conn, err := grpc.Dial(hostname+":"+l.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logrus.Errorf("cannot connect to %s leaderServer: %v", hostname, err)
			}
			defer conn.Close()
			client := pb.NewLeaderServerClient(conn)
			_, err = client.SetLeader(context.Background(), &pb.SetLeaderRequest{Leader: leader})
			if err != nil {
				logrus.Errorf("failed to set leader of %s: %v", hostname, err)
			}
		}(member)
	}
	wg.Wait()
	return
}

// electLeader elects the leader and returns the leader hostname.
func (l *LeaderServer) electLeaderFromMembers(members map[string]*membership.Member) string {
	// Find the members with the largest heartbeat, and then find the member with the smallest hostname
	maxHeartbeat := 0
	hostname := ""
	for _, member := range members {
		if member.Heartbeat > maxHeartbeat {
			maxHeartbeat = member.Heartbeat
			hostname = member.GetName()
		} else if member.Heartbeat == maxHeartbeat {
			if member.GetName() < hostname {
				hostname = member.GetName()
			}
		}
	}
	return hostname
}

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
