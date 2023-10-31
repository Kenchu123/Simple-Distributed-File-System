package leaderserver

import (
	"context"
	"fmt"

	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/metadata"
)

func (l *LeaderServer) GetBlockInfo(ctx context.Context, in *pb.GetBlockInfoRequest) (*pb.GetBlockInfoReply, error) {
	err := l.acquireFileSemaphore(in.FileName)
	if err != nil {
		return nil, err
	}
	blockInfo, err := l.getBlockInfo(in.FileName)
	if err != nil {
		return nil, err
	}
	getBlockInfoReplyBlockMeta := map[int64]*pb.GetBlockInfoReply_BlockMeta{}
	for blockID, blockMeta := range blockInfo {
		getBlockInfoReplyBlockMeta[blockID] = &pb.GetBlockInfoReply_BlockMeta{
			HostNames: blockMeta.HostNames,
			FileName:  blockMeta.FileName,
			BlockID:   blockMeta.BlockID,
		}
	}
	return &pb.GetBlockInfoReply{
		BlockInfo: getBlockInfoReplyBlockMeta,
	}, nil
}

func (l *LeaderServer) acquireFileSemaphore(fileName string) error {
	if _, ok := l.fileSemaphore[fileName]; !ok {
		return fmt.Errorf("file %s not found", fileName)
	}
	l.fileSemaphore[fileName].Acquire(context.Background(), 1)
	return nil
}

func (l *LeaderServer) getBlockInfo(fileName string) (metadata.BlockInfo, error) {
	return l.metadata.GetBlockInfo(fileName)
}

func (l *LeaderServer) GetFileOK(ctx context.Context, in *pb.GetFileOKRequest) (*pb.GetFileOKReply, error) {
	if _, ok := l.fileSemaphore[in.FileName]; !ok {
		return nil, fmt.Errorf("file %s not found", in.FileName)
	}
	l.releaseFileSemaphore(in.FileName, 1)
	return &pb.GetFileOKReply{}, nil
}

func (l *LeaderServer) releaseFileSemaphore(fileName string, weight int64) {
	l.fileSemaphore[fileName].Release(weight)
}
