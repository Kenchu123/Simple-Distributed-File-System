package leaderserver

import (
	"context"
	"fmt"
	"sync"

	pb "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"golang.org/x/sync/semaphore"
)

// FileLock is a lock for file
type FileLock struct {
	fileSempahore map[string]*semaphore.Weighted
	mu            sync.Mutex
}

// NewFileLock returns a new FileLock
func NewFileLock() *FileLock {
	return &FileLock{
		fileSempahore: map[string]*semaphore.Weighted{},
		mu:            sync.Mutex{},
	}
}

// AcquireReadLock acquires a read lock for a file through gRPC.
func (l *LeaderServer) AcquireReadLock(ctx context.Context, in *pb.AcquireLockRequest) (*pb.AcquireLockReply, error) {
	return &pb.AcquireLockReply{}, l.fileLock.acquireLock(in.GetFileName(), 1)
}

// ReleaseReadLock releases a read lock for a file through gRPC.
func (l *LeaderServer) ReleaseReadLock(ctx context.Context, in *pb.ReleaseLockRequest) (*pb.ReleaseLockReply, error) {
	return &pb.ReleaseLockReply{}, l.fileLock.releaseLock(in.GetFileName(), 1)
}

// AcquireWriteLock acquires a write lock for a file through gRPC.
func (l *LeaderServer) AcquireWriteLock(ctx context.Context, in *pb.AcquireLockRequest) (*pb.AcquireLockReply, error) {
	return &pb.AcquireLockReply{}, l.fileLock.acquireLock(in.GetFileName(), 2)
}

// ReleaseWriteLock releases a write lock for a file through gRPC.
func (l *LeaderServer) ReleaseWriteLock(ctx context.Context, in *pb.ReleaseLockRequest) (*pb.ReleaseLockReply, error) {
	return &pb.ReleaseLockReply{}, l.fileLock.releaseLock(in.GetFileName(), 2)
}

// acquireLock acquires a lock for a file with weight.
func (fl *FileLock) acquireLock(fileName string, weight int64) error {
	if _, ok := fl.fileSempahore[fileName]; !ok {
		fl.mu.Lock()
		fl.fileSempahore[fileName] = semaphore.NewWeighted(2)
		fl.mu.Unlock()
	}
	fl.fileSempahore[fileName].Acquire(context.Background(), weight)
	return nil
}

// releaseLock releases a lock for a file with weight.
func (fl *FileLock) releaseLock(fileName string, weight int64) error {
	if _, ok := fl.fileSempahore[fileName]; !ok {
		return fmt.Errorf("file %s lock not found", fileName)
	}
	fl.fileSempahore[fileName].Release(weight)
	return nil
}
