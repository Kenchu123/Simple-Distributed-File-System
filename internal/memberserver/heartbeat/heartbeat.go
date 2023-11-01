package heartbeat

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/membership"
)

type Heartbeat struct {
	Config     *config.Config
	IsRunning  bool
	Membership *membership.Membership

	// ticker
	heartbeatTicker     *time.Ticker
	heartbeatTickerDone chan bool

	// udp
	udpServer *UdpServer

	// failure detector
	failureDetectTicker     *time.Ticker
	failureDetectTickerDone chan bool

	// cleanup
	cleanupTicker     *time.Ticker
	cleanupTickerDone chan bool
}

var lock = &sync.Mutex{}
var instance *Heartbeat

// GetInstance returns the singleton instance of heartbeat
func GetInstance() (*Heartbeat, error) {
	lock.Lock()
	defer lock.Unlock()
	var err error
	if instance == nil {
		instance, err = New()
		if err != nil {
			return nil, err
		}
	}
	return instance, nil
}

// New creates a new heartbeat
func New() (*Heartbeat, error) {
	conf, err := config.GetInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new heartbeat: %w", err)
	}
	return &Heartbeat{
		Config:                  conf,
		IsRunning:               false,
		Membership:              nil,
		heartbeatTicker:         nil,
		heartbeatTickerDone:     make(chan bool),
		udpServer:               nil,
		failureDetectTicker:     nil,
		failureDetectTickerDone: make(chan bool),
	}, nil
}

// Start starts the heartbeat
func (h *Heartbeat) Start() {
	var err error
	h.Membership, err = membership.New(h.Config.Heartbeat.Introducer, h.Config.Heartbeat.TargetNumber)
	if err != nil {
		logrus.Errorf("failed to start: %v", err)
		return
	}
	h.udpServer, err = NewUdpServer(h.Config.Heartbeat.Port)
	if err != nil {
		logrus.Errorf("failed to start: %v", err)
		return
	}
	h.IsRunning = true
	go h.startHeartbeating()
	go h.startReceiving()
	go h.startDetectingFailure()
	go h.startCleaningUp()
}

func (h *Heartbeat) startHeartbeating() {
	logrus.Info("Start heartbeating")
	h.heartbeatTicker = time.NewTicker(h.Config.Heartbeat.Interval)
	defer h.heartbeatTicker.Stop()
	for {
		select {
		case <-h.heartbeatTickerDone:
			return
		case <-h.heartbeatTicker.C:
			h.sendHeartbeat()
		}
	}
}

func (h *Heartbeat) sendHeartbeat() {
	h.Membership.IncreaseSelfHeartbeat()
	hostnames := h.Membership.GetHeartbeatTargetMembers(h.Config.Machines)
	logrus.Debug("Heartbeat target members: ", hostnames)
	for _, hostname := range hostnames {
		go func(hostname string) {
			client, err := NewUdpClient(hostname, h.Config.Heartbeat.Port)
			if err != nil {
				logrus.Errorf("failed to create udp client: %v", err)
				return
			}
			payload, err := membership.Serialize(h.Membership)
			if err != nil {
				logrus.Errorf("failed to serialize membership: %v", err)
				return
			}
			_, err = client.Send(payload)
			if err != nil {
				logrus.Errorf("Failed to send heartbeat to %s, error: %v", hostname, err)
				return
			}
			logrus.Debugf("Sending heartbeat to %s: %s", hostname, h.Membership)
		}(hostname)
	}
}

func (h *Heartbeat) startReceiving() {
	logrus.Info("Start receiving heartbeat")
	h.udpServer.Serve(h.receiveHeartbeat)
}

func (h *Heartbeat) receiveHeartbeat(addr net.Addr, buffer []byte) {
	if h.Config.Heartbeat.DropRate > 0 {
		rand := rand.Float32()
		if rand < h.Config.Heartbeat.DropRate {
			logrus.Debugf("Dropping heartbeat from %s", addr.String())
			return
		}
	}
	membership, err := membership.Deserialize(buffer)
	if err != nil {
		logrus.Errorf("failed to deserialize membership from %s: %v", addr.String(), err)
		return
	}
	logrus.Debugf("Received heartbeat from %s: %s", addr.String(), membership)
	h.Membership.Update(membership)
}

func (h *Heartbeat) startDetectingFailure() {
	logrus.Info("Start detecting failure")
	h.failureDetectTicker = time.NewTicker(h.Config.FailureDetect.Interval)
	defer h.failureDetectTicker.Stop()
	for {
		select {
		case <-h.failureDetectTickerDone:
			return
		case <-h.failureDetectTicker.C:
			h.detectFailure()
		}
	}
}

func (h *Heartbeat) detectFailure() {
	logrus.Debug("Detecting failure")
	h.Membership.DetectFailure(h.Config.FailureDetect)
}

func (h *Heartbeat) startCleaningUp() {
	logrus.Info("Start cleaning up Membership")
	h.cleanupTicker = time.NewTicker(h.Config.Cleanup.Interval)
	defer h.cleanupTicker.Stop()
	for {
		select {
		case <-h.cleanupTickerDone:
			return
		case <-h.cleanupTicker.C:
			h.Membership.CleanUp(h.Config.Cleanup.Timeout)
		}
	}
}

// Stop stops the heartbeat
func (h *Heartbeat) Stop() {
	h.IsRunning = false
	go h.stopHeartbeating()
	go h.stopReceiving()
	go h.stopDetectingFailure()
	go h.stopCleaningUp()
}

func (h *Heartbeat) stopHeartbeating() {
	logrus.Info("Stop heartbeating")
	h.heartbeatTickerDone <- true
}

func (h *Heartbeat) stopReceiving() {
	logrus.Info("Stop receiving heartbeat")
	h.udpServer.Stop()
}

func (h *Heartbeat) stopDetectingFailure() {
	logrus.Info("Stop detecting failure")
	h.failureDetectTickerDone <- true
}

func (h *Heartbeat) stopCleaningUp() {
	logrus.Info("Stop cleaning up Membership")
	h.cleanupTickerDone <- true
}

// SetSuspicion sets suspicion enabled
func (h *Heartbeat) SetSuspicion(enabled bool) {
	logrus.Info("[SUSPICION] Set suspicion enabled to ", enabled)
	h.Config.FailureDetect.Suspicion.Enabled = enabled
}

// SetDropRate sets drop rate
func (h *Heartbeat) SetDropRate(dropRate float32) {
	logrus.Info("[DROPRATE] Set drop rate to ", dropRate)
	h.Config.Heartbeat.DropRate = dropRate
}
