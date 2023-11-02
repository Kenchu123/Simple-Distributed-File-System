package leaderserver

import (
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
)

type ToReplicates []ToReplicate

type ToReplicate struct {
	FileName string
	BlockID  int64
	From     string
	To       string
}

func (l *LeaderServer) startRecoveringReplica() {
	logrus.Info("Start recovering replica")
	l.recoverReplicaTicker = time.NewTicker(time.Second * 5)
	defer l.recoverReplicaTicker.Stop()
	for {
		select {
		case <-l.recoverReplicaTickerDone:
			return
		case <-l.recoverReplicaTicker.C:
			l.recoverReplica()
		}
	}
}

func (l *LeaderServer) recoverReplica() {
	// TODO: only leader runs this
	heartbeat, err := heartbeat.GetInstance()
	if err != nil {
		logrus.Errorf("Failed to get heartbeat instance: %v", err)
		return
	}
	membership := heartbeat.GetMembership()
	if membership == nil {
		logrus.Errorf("Failed to get membership instance")
		return
	}
	members := membership.GetAliveMembers()
	hostnames := make([]string, 0, len(members))
	hostnameSet := make(map[string]struct{}, len(members))
	for _, member := range members {
		hostnames = append(hostnames, member.GetName())
		hostnameSet[member.GetName()] = struct{}{}
	}
	// TODO: require the metadata lock
	// scan all file blocks and check if the replica hostname is in the member list
	toReclicates := make(ToReplicates, 0)
	for fileName, blockInfo := range l.metadata.GetFileInfo() {
		for blockID, blockMeta := range blockInfo {
			alivedHostnames := []string{}
			alivedHostnamesSet := make(map[string]struct{})
			for _, hostname := range blockMeta.HostNames {
				if _, ok := hostnameSet[hostname]; ok {
					alivedHostnames = append(alivedHostnames, hostname)
					alivedHostnamesSet[hostname] = struct{}{}
				}
			}
			l.metadata.AddOrUpdateBlockMeta(fileName, metadata.BlockMeta{
				HostNames: alivedHostnames,
				FileName:  fileName,
				BlockID:   blockID,
			})

			// if the alivedHostnames is less than replicaFactor, randomly select the hostname from the member list
			newHostnames := make(map[string]struct{})
			if len(alivedHostnamesSet) < l.replicationFactor {
				for i := len(alivedHostnamesSet); i < l.replicationFactor; i++ {
					var randomHostname string
					for {
						randomHostname = hostnames[rand.Intn(len(hostnames))]
						if _, ok := alivedHostnamesSet[randomHostname]; ok {
							continue
						}
						if _, ok := newHostnames[randomHostname]; ok {
							continue
						}
						newHostnames[randomHostname] = struct{}{}
						break
					}
				}
			}
			for hostname := range newHostnames {
				toReclicates = append(toReclicates, ToReplicate{
					FileName: fileName,
					BlockID:  blockID,
					From:     alivedHostnames[rand.Intn(len(alivedHostnames))],
					To:       hostname,
				})
			}
		}
	}

	if len(toReclicates) == 0 {
		return
	}

	logrus.Infof("To reclicate %+v", toReclicates)
	var wg sync.WaitGroup
	for _, toReplicate := range toReclicates {
		wg.Add(1)
		go func(toReplicate ToReplicate) {
			defer wg.Done()
			err := l.replicate(toReplicate)
			if err != nil {
				logrus.Errorf("Failed to replicate %+v: %v", toReplicate, err)
				return
			}
			logrus.Infof("Replicated %+v", toReplicate)
			// update metadata
			blockInfo, err := l.metadata.GetBlockMeta(toReplicate.FileName, toReplicate.BlockID)
			if err != nil {
				logrus.Errorf("Failed to get block meta %+v: %v", toReplicate, err)
				return
			}
			l.metadata.AddOrUpdateBlockMeta(toReplicate.FileName, metadata.BlockMeta{
				HostNames: append(blockInfo.HostNames, toReplicate.To),
				FileName:  toReplicate.FileName,
				BlockID:   toReplicate.BlockID,
			})
		}(toReplicate)
	}
	wg.Wait()
	return
}

func (l *LeaderServer) replicate(toReplicate ToReplicate) error {
	logrus.Infof("Replicating %+v", toReplicate)
	// sleep for a 3 seconds
	time.Sleep(time.Second * 3)
	return nil
}
