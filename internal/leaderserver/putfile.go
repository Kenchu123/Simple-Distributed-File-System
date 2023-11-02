package leaderserver

import (
	"context"
	"math/rand"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/memberserver/heartbeat"
	"golang.org/x/sync/semaphore"
)

// PutBlockInfo handles the request to choose the block to put the file
func (l *LeaderServer) PutBlockInfo(ctx context.Context, in *pb.PutBlockInfoRequest) (*pb.PutBlockInfoReply, error) {
	// acquire file semaphore if file exists
	if l.metadata.IsFileExist(in.FileName) {
		l.acquireFileSemaphore(in.FileName, 2)
	}
	blockInfo, err := l.putBlockInfo(in.FileName, in.FileSize)
	if err != nil {
		return nil, err
	}
	putBlockInfoReplyBlockMeta := map[int64]*pb.BlockMeta{}
	for blockID, blockMeta := range blockInfo {
		putBlockInfoReplyBlockMeta[blockID] = &pb.BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	return &pb.PutBlockInfoReply{
		BlockInfo: putBlockInfoReplyBlockMeta,
	}, nil
}

// putBlockInfo select the block to put the file
func (l *LeaderServer) putBlockInfo(fileName string, fileSize int64) (metadata.BlockInfo, error) {
	blocksNum := fileSize / l.blockSize
	if fileSize%l.blockSize != 0 {
		blocksNum++
	}
	blockInfo := map[int64]metadata.BlockMeta{}
	for i := int64(0); i < blocksNum; i++ {
		blockInfo[i] = metadata.BlockMeta{
			HostNames: l.selectBlockHosts(),
			FileName:  fileName,
			BlockID:   i,
		}
	}
	return blockInfo, nil
}

// selectBlockHosts randomly select replicationFactor hosts
func (l *LeaderServer) selectBlockHosts() []string {
	heartbeat, err := heartbeat.GetInstance()
	if err != nil {
		panic(err)
	}
	// Assuming that the alive members are more than replicationFactor
	membership := heartbeat.GetMembership()
	if membership == nil {
		logrus.Errorf("failed to get membership instance")
		return []string{}
	}
	members := membership.GetAliveMembers()
	hostnames := []string{}
	for _, member := range members {
		hostnames = append(hostnames, member.GetName())
	}
	// randomly select replicationFactor hosts
	rand.Shuffle(len(hostnames), func(i, j int) { hostnames[i], hostnames[j] = hostnames[j], hostnames[i] })
	return hostnames[:l.replicationFactor]
}

func (l *LeaderServer) PutFileOK(ctx context.Context, in *pb.PutFileOKRequest) (*pb.PutFileOKReply, error) {
	// release the file semaphore if file exists
	if l.metadata.IsFileExist(in.FileName) {
		l.releaseFileSemaphore(in.FileName, 2)
	}
	blockInfo := metadata.BlockInfo{}
	for blockID, blockMeta := range in.BlockInfo {
		blockInfo[blockID] = metadata.BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	l.metadata.AddOrUpdateFile(in.FileName, blockInfo)
	l.fileSemaphore[in.FileName] = semaphore.NewWeighted(2)
	return &pb.PutFileOKReply{}, nil
}
