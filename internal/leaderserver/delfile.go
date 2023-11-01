package leaderserver

import (
	"context"
	"fmt"

	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
)

func (l *LeaderServer) DelFile(ctx context.Context, in *pb.DelFileRequest) (*pb.DelFileReply, error) {
	if !l.metadata.IsFileExist(in.FileName) {
		return nil, fmt.Errorf("file %s does not exist", in.FileName)
	}
	// TODO: acquire file semaphore?
	l.metadata.DelFile(in.FileName)
	delete(l.fileSemaphore, in.FileName)
	return &pb.DelFileReply{}, nil
}
