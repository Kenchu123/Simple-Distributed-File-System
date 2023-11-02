package command

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/command/proto"
	client "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/sdfsclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CommandClient handles commands to sdfsclient.
type CommandClient struct {
	hostname string
	port     string
}

// NewCommandClient creates a new command client.
func NewCommandClient(hostname, port string) *CommandClient {
	return &CommandClient{
		hostname: hostname,
		port:     port,
	}
}

// ExecuteCommand executes a command through gRPC.
func (c *CommandClient) ExecuteCommand(command string, args []string) (string, error) {
	conn, err := grpc.Dial(c.hostname+":"+c.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatal(fmt.Errorf("cannot dial command server %s: %v", c.hostname, err))
	}
	defer conn.Close()

	client := pb.NewCommandClient(conn)
	r, err := client.ExecuteCommand(context.Background(), &pb.ExecuteCommandRequest{
		Command: command,
		Args:    args,
	})
	if err != nil {
		return "", err
	}
	return r.GetOutput(), nil
}

// CommandServer handles commands from sdfsclient.
type CommandServer struct {
	port       string
	configPath string

	pb.UnimplementedCommandServer
}

// NewCommandServer creates a new command server.
func NewCommandServer(port, configPath string) *CommandServer {
	return &CommandServer{
		port:       port,
		configPath: configPath,
	}
}

// Run runs the command server.
func (s *CommandServer) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		logrus.Fatalf("failed to listen on port %s: %v\n", s.port, err)
		return
	}
	defer listen.Close()
	grpcServer := grpc.NewServer()
	pb.RegisterCommandServer(grpcServer, s)
	logrus.Infof("CommandServer listening on port %s", s.port)
	if err := grpcServer.Serve(listen); err != nil {
		logrus.Fatalf("failed to serve: %v\n", err)
		return
	}
	return
}

// ExecuteCommand executes a command though gRPC.
func (s *CommandServer) ExecuteCommand(ctx context.Context, in *pb.ExecuteCommandRequest) (*pb.ExecuteCommandReply, error) {
	output, err := s.executeCommand(in.Command, in.Args)
	if err != nil {
		return nil, err
	}
	return &pb.ExecuteCommandReply{
		Output: output,
	}, nil
}

// ExecuteCommand executes a command.
func (s *CommandServer) executeCommand(command string, args []string) (string, error) {
	sdfsClient, err := client.NewClient(s.configPath)
	if err != nil {
		return "", err
	}
	switch command {
	case "put":
		if len(args) != 2 {
			return "", fmt.Errorf("put command requires 2 arguments")
		}
		return "", sdfsClient.PutFile(args[0], args[1])
	case "get":
		if len(args) != 2 {
			return "", fmt.Errorf("get command requires 2 arguments")
		}
		return "", sdfsClient.GetFile(args[0], args[1])
	case "delete":
		if len(args) != 1 {
			return "", fmt.Errorf("delete command requires 1 argument")
		}
		return "", sdfsClient.DelFile(args[0])
	case "ls":
		if len(args) != 1 {
			return "", fmt.Errorf("ls command requires 1 argument")
		}
		return sdfsClient.LsFile(args[0])
	case "store":
		if len(args) != 0 {
			return "", fmt.Errorf("store command requires 0 argument")
		}
		return sdfsClient.Store()
	default:
		return "", fmt.Errorf("unknown command %s", command)
	}
}
