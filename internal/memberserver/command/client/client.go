package client

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/command/socket"
)

type Client struct {
	machines         []config.Machine
	memberServerPort string
}

type Result struct {
	Hostname string
	Message  string
	Err      error
}

// New creates a new client
func New(configPath string, machineRegex string) (*Client, error) {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	machines, err := conf.FilterMachines(machineRegex)
	if err != nil {
		return nil, err
	}
	return &Client{
		machines:         machines,
		memberServerPort: conf.MemberServerPort,
	}, nil
}

// Run runs the client
func (c *Client) Run(args []string) map[string]Result {
	var wg = &sync.WaitGroup{}
	result := make(chan Result)
	for _, machine := range c.machines {
		wg.Add(1)
		go func(machine config.Machine) {
			defer wg.Done()
			response, err := sendCommand(machine.Hostname, c.memberServerPort, strings.Join(args, " "))
			if err != nil {
				result <- Result{
					Hostname: machine.Hostname,
					Message:  "",
					Err:      err,
				}
				return
			}
			result <- Result{
				Hostname: machine.Hostname,
				Message:  response,
				Err:      nil,
			}
		}(machine)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	// combine all the results
	var results map[string]Result = map[string]Result{}
	for r := range result {
		results[r.Hostname] = r
	}
	return results
}

// sendCommand sends a command to a server
func sendCommand(hostname string, port string, msg string) (string, error) {
	if len(msg) == 0 {
		return "", fmt.Errorf("empty message")
	}
	conn, err := newConnection(hostname, port)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	socket.Send(conn, []byte(msg))
	_, response, err := socket.Receive(conn)
	if err != nil {
		return "", fmt.Errorf("failed to receive response: %w", err)
	}
	return string(response), nil
}

// newConnection creates a new socket connection
func newConnection(hostname string, port string) (net.Conn, error) {
	d := net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, err := d.Dial("tcp", hostname+":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s:%s: %w", hostname, port, err)
	}
	return conn, nil
}
