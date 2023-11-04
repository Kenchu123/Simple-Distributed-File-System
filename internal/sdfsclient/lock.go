package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	leaderServerProto "gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/leaderserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// acquireFileReadLock gets the read lock of a file.
func (c *Client) acquireFileReadLock(leader, fileName string) error {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	// TODO: acquire lock timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	_, err = client.AcquireReadLock(ctx, &leaderServerProto.AcquireLockRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to acquire read lock: %v", err)
	}
	c.fileReadLocks[leader+":"+fileName] = true
	c.releaseFileReadLockOnInterrupt(leader, fileName)
	return nil
}

// releaseLockOnInterrupt releases the lock on interrupt.
func (c *Client) releaseFileReadLockOnInterrupt(leader, fileName string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		if _, ok := c.fileReadLocks[leader+":"+fileName]; !ok {
			os.Exit(1)
		}
		logrus.Infof("Release read lock on interrupt for file %s", fileName)
		c.releaseFileReadLock(leader, fileName)
		os.Exit(1)
	}()
}

// releaseFileReadLock gets the read lock of a file.
func (c *Client) releaseFileReadLock(leader, fileName string) error {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	_, err = client.ReleaseReadLock(ctx, &leaderServerProto.ReleaseLockRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to release read lock: %v", err)
	}
	delete(c.fileReadLocks, leader+":"+fileName)
	return nil
}

// acquireFileWriteLock gets the read lock of a file.
func (c *Client) acquireFileWriteLock(leader, fileName string) error {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	// TODO: acquire lock timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	_, err = client.AcquireWriteLock(ctx, &leaderServerProto.AcquireLockRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to acquire write lock: %v", err)
	}
	c.fileWriteLocks[leader+":"+fileName] = true
	c.releaseFileWriteLockOnInterrupt(leader, fileName)
	return nil
}

// releaseLockOnInterrupt releases the lock on interrupt.
func (c *Client) releaseFileWriteLockOnInterrupt(leader, fileName string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		if _, ok := c.fileWriteLocks[leader+":"+fileName]; !ok {
			os.Exit(1)
		}
		logrus.Infof("Release write lock on interrupt for file %s", fileName)
		c.releaseFileWriteLock(leader, fileName)
		os.Exit(1)
	}()
}

// releaseFileWriteLock gets the read lock of a file.
func (c *Client) releaseFileWriteLock(leader, fileName string) error {
	conn, err := grpc.Dial(leader+":"+c.leaderServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("cannot connect to %s leaderServer: %v", leader, err)
	}
	defer conn.Close()

	client := leaderServerProto.NewLeaderServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	_, err = client.ReleaseWriteLock(ctx, &leaderServerProto.ReleaseLockRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to release write lock: %v", err)
	}
	delete(c.fileWriteLocks, leader+":"+fileName)
	return nil
}
