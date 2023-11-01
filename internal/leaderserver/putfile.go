package leaderserver

import (
	"context"

	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/metadata"
	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
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
	// TODO: choose the node to put th file block
	return map[int64]metadata.BlockMeta{
		0: {
			HostNames: []string{"localhost"},
			FileName:  fileName,
			BlockID:   0,
		},
		1: {
			HostNames: []string{"localhost"},
			FileName:  fileName,
			BlockID:   1,
		},
	}, nil
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
