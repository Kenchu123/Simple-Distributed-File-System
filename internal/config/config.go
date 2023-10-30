package config

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for the servers
type Config struct {
	Machines         []Machine `yaml:"machines"`
	LeaderServerPort string    `yaml:"leader_server_port"`
	DataServerPort   string    `yaml:"data_server_port"`
	BlocksDir        string    `yaml:"blocks_dir"`
}

// Machine is the configuration for a single server
type Machine struct {
	Hostname string `yaml:"hostname"`
	ID       string `yaml:"id"`
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
