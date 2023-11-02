package leaderserver

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (l *LeaderServer) startSyncingMetadata() {
	logrus.Info("Start syncing metadata")
	l.syncMetadataTicker = time.NewTicker(time.Second * 5)
	defer l.syncMetadataTicker.Stop()
	for {
		select {
		case <-l.syncMetadataTickerDone:
			return
		case <-l.syncMetadataTicker.C:
			l.syncMetadata()
		}
	}
}

func (l *LeaderServer) stopSyncingMetadata() {
	l.syncMetadataTickerDone <- true
}

func (l *LeaderServer) syncMetadata() {
	// Step1: get leader
	// only leader don't need to sync metadata
	leader := l.getLeader()
	if leader == l.hostname {
		return
	}
	if leader == "" {
		logrus.Infof("No leader, skip syncing metadata")
		return
	}
	// Step2: get metadata from leader
	conn, err := grpc.Dial(leader+":"+l.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Errorf("cannot connect to %s leaderServer: %v", leader, err)
		return
	}
	defer conn.Close()
	client := pb.NewLeaderServerClient(conn)
	r, err := client.GetMetadata(context.Background(), &pb.GetMetadataRequest{})
	if err != nil {
		logrus.Errorf("cannot get metadata from %s leaderServer: %v", leader, err)
		return
	}
	// Step3: update metadata
	fileInfo := r.GetFileInfo()
	for fileName, blockInfo := range fileInfo {
		newBlockInfo := metadata.BlockInfo{}
		for blockID, blockMeta := range blockInfo.GetBlockInfo() {
			newBlockInfo[blockID] = metadata.BlockMeta{
				BlockID:   blockMeta.GetBlockID(),
				HostNames: blockMeta.GetHostNames(),
				FileName:  blockMeta.GetFileName(),
			}
		}
		l.metadata.AddOrUpdateBlockInfo(fileName, newBlockInfo)
	}
}
