package server

import (
	"sync"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/dataserver"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver"
)

// Server handles file operations to SDFS.
type Server struct {
	*leaderserver.LeaderServer
	*dataserver.DataServer
}

// NewServer creates a new Server.
func NewServer(configPath string) (*Server, error) {
	config, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	leaderServer := leaderserver.NewLeaderServer(config.LeaderServerPort)
	dataServer := dataserver.NewDataServer(config.DataServerPort, config.BlocksDir)
	return &Server{
		LeaderServer: leaderServer,
		DataServer:   dataServer,
	}, nil
}

// Run starts the server.
func (s *Server) Run() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.LeaderServer.Run()
	}()
	go func() {
		defer wg.Done()
		s.DataServer.Run()
	}()
	wg.Wait()
}
