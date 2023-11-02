package leaderserver

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/membership"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
		logrus.Debugf("failed to get membership instance")
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
