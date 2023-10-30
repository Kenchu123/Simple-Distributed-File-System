package leaderserver

import "github.com/sirupsen/logrus"

// LeaderServer handles file operations permission and Leader election.
type LeaderServer struct {
	port   string
	leader string
}

// NewLeader creates a new Leader.
func NewLeaderServer(port string) *LeaderServer {
	return &LeaderServer{
		port:   port,
		leader: "This is a test leader hostname hayeyee",
	}
}

// Run starts the Leader.
func (l *LeaderServer) Run() {
	logrus.Info("Leader server is running on port ", l.port)
}

func (l *LeaderServer) GetLeader() string {
	return l.leader + ":" + l.port
}

func (l *LeaderServer) SetLeader(leader string) {
	l.leader = leader
}

// TODO: Elect Leader, Put file, get file, del file
