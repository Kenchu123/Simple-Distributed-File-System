package config

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for the servers
type Config struct {
	Machines         []Machine     `yaml:"machines"`
	LeaderServerPort string        `yaml:"leader_server_port"`
	DataServerPort   string        `yaml:"data_server_port"`
	MemberServerPort string        `yaml:"member_server_port"`
	BlocksDir        string        `yaml:"blocks_dir"`
	BlockSize        int64         `yaml:"block_size"`
	RelicationFactor int           `yaml:"replication_factor"`
	Heartbeat        Heartbeat     `yaml:"heartbeat"`
	FailureDetect    FailureDetect `yaml:"failure_detect"`
	Cleanup          Cleanup       `yaml:"cleanup"`
}

// Machine is the configuration for a single server
type Machine struct {
	Hostname string `yaml:"hostname"`
	ID       string `yaml:"id"`
}

type Heartbeat struct {
	Port         string        `yaml:"port"`         // UDP port for heartbeat
	Interval     time.Duration `yaml:"interval"`     // send heartbeat every <interval> millisecond
	DropRate     float32       `yaml:"drop_rate"`    // DropRate of the udp packet
	Introducer   string        `yaml:"introducer"`   // Introducer's hostname
	TargetNumber int           `yaml:"targetNumber"` // number of target machines
}

type FailureDetect struct {
	Interval       time.Duration `yaml:"interval"`        // check failure every <interval> millisecond
	FailureTimeout time.Duration `yaml:"failure_timeout"` // set to failure if no heartbeat received for <timeout> millisecond
	Suspicion      struct {
		Enabled        bool          `yaml:"enabled"`         // enable suspicion
		SuspectTimeout time.Duration `yaml:"suspect_timeout"` // set to suspect if no heartbeat received for <timeout> millisecond
		FailureTimeout time.Duration `yaml:"failure_timeout"` // set to failure if suspected and no heartbeat received for <timeout> millisecond
	} `yaml:"suspicion"` // suspicion configuration
}

type Cleanup struct {
	Interval time.Duration `yaml:"interval"` // clean up left / failure every <interval> millisecond
	Timeout  time.Duration `yaml:"timeout"`  // remove from membership if left or failed for <timeout> millisecond
}

var lock = &sync.Mutex{}
var instance *Config = nil

// GetInstance returns the instance of the config
func GetInstance() (*Config, error) {
	lock.Lock()
	defer lock.Unlock()
	if instance == nil {
		return nil, fmt.Errorf("config is not initialized")
	}
	return instance, nil
}

// New reads the configuration file and returns the configuration
func NewConfig(path string) (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	instance = config
	return config, nil
}

// FilterMachines filters the machines based on the regex
func (c *Config) FilterMachines(regex string) ([]Machine, error) {
	var machines []Machine
	reg, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	for _, machine := range c.Machines {
		if reg.MatchString(machine.Hostname) {
			machines = append(machines, machine)
		}
	}
	return machines, nil
}
