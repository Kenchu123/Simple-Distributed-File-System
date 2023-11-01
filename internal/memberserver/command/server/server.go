package server

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/handler"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/socket"
)

var chunkSize = 4096

// Server handles server
type Server struct {
	server net.Listener
}

// New creates a new server
func New(configPath, port string) (*Server, error) {
	_, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	server, err := net.Listen("tcp", hostName+":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s:%s: %w", hostName, port, err)
	}
	logrus.Infof("Listening on %s:%s", hostName, port)
	return &Server{
		server: server,
	}, nil
}

// NewMemberServer creates a new server
func NewMemberServer(port string) *Server {
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logrus.Fatalf("failed to listen on port %s: %v\n", port, err)
	}
	logrus.Infof("MemberServer listening on port %s", port)
	return &Server{
		server: server,
	}
}

// Run runs the server
func (s *Server) Run() {
	defer s.close()
	for {
		conn, err := s.accept()
		if err != nil {
			logrus.Errorf("failed to accept new connection: %v", err)
			continue
		}
		go handleRecieveCommand(conn)
	}
}

// Close closes the server
func (s *Server) close() {
	s.server.Close()
}

// Accept accepts a new connection
func (s *Server) accept() (net.Conn, error) {
	conn, err := s.server.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept new connection: %w", err)
	}
	return conn, nil
}

// handleConnection handles a new connection
func handleRecieveCommand(conn net.Conn) {
	defer conn.Close()
	_, content, err := socket.Receive(conn)
	if err != nil {
		logrus.Errorf("failed to receive message: %v", err)
		return
	}

	handler := handler.NewRootHandler()
	result, err := handler.Handle(strings.Split(string(content), " "))
	if err != nil {
		logrus.Error(err)
		conn.Write([]byte(err.Error()))
		return
	}
	if _, err = socket.Send(conn, []byte(result)); err != nil {
		logrus.Errorf("failed to send message: %v", err)
		return
	}
}
