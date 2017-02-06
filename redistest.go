package redistest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/redis.v5"
)

type ServerType int

const (
	Master ServerType = iota
	Slave
	Sentinel
)

const redisServerExe = "redis-server"

const (
	DefaultMasterPort     = 6379
	DefaultSlavePort      = 6380
	DefaultSentinelPort   = 26379
	DefaultSentinelMaster = "mymaster"
)

type Server struct {
	cmd        *exec.Cmd
	currentCmd exec.Cmd
	port       int
	done       chan bool
}

func RunServer(port int) (*Server, error) {
	cmd := exec.Command(redisServerExe,
		"--port", fmt.Sprintf("%d", port),
		"--dir", "/tmp",
		"--dbfilename", fmt.Sprintf("redis_test.%d.%d.rdb", port, time.Now().UnixNano()),
	)
	return runServer(port, cmd)
}

func (s *Server) RunSlaveServer(port int) (*Server, error) {
	cmd := exec.Command(redisServerExe,
		"--port", fmt.Sprintf("%d", port),
		fmt.Sprintf("--slaveof localhost %d", s.port),
		"--dir", "/tmp",
		"--dbfilename", fmt.Sprintf("redis_test.%d.%d.rdb", port, time.Now().UnixNano()),
	)
	return runServer(port, cmd)
}

func (s *Server) RunSentinelServer(port int, masterName string) (*Server, error) {
	cmd := exec.Command(redisServerExe,
		os.DevNull,
		"--port", fmt.Sprintf("%d", port),
		"--sentinel",
	)
	sentinel, err := runServer(port, cmd)
	if err != nil {
		return nil, err
	}

	c := sentinel.NewClient()
	defer c.Close()

	for _, cmd := range []*redis.StatusCmd{
		redis.NewStatusCmd("SENTINEL", "MONITOR", masterName, "127.0.0.1", s.port, "1"),
		redis.NewStatusCmd("SENTINEL", "SET", masterName, "down-after-milliseconds", "500"),
		redis.NewStatusCmd("SENTINEL", "SET", masterName, "failover-timeout", "1000"),
		redis.NewStatusCmd("SENTINEL", "SET", masterName, "parallel-syncs", "1"),
	} {
		c.Process(cmd)
		if err := cmd.Err(); err != nil {
			sentinel.Stop()
			return nil, err
		}
	}

	return sentinel, err
}

func runServer(port int, cmd *exec.Cmd) (*Server, error) {
	s := &Server{
		cmd:  cmd,
		port: port,
	}

	if err := s.Run(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) IsRunning() bool {
	c := s.NewClient()
	defer c.Close()

	pong, _ := c.Ping().Result()
	return pong == "PONG"
}

func (s *Server) Run() error {
	if s.IsRunning() {
		return errors.New("Redis server is already running")
	}

	s.currentCmd = *s.cmd

	if err := s.currentCmd.Start(); err != nil {
		return err
	}

	s.done = make(chan bool, 1)

	var err error
	go func() {
		err = s.currentCmd.Wait()
		s.done <- true
	}()

	time.Sleep(100 * time.Millisecond)

	return err
}

func (s *Server) Stop() error {
	if err := s.currentCmd.Process.Kill(); err != nil {
		return err
	}
	<-s.done
	return nil
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Addr() string {
	return fmt.Sprintf("localhost:%d", s.port)
}

func (s *Server) Flush() error {
	c := s.NewClient()
	defer c.Close()

	return c.FlushAll().Err()
}

func (s *Server) NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: fmt.Sprintf(":%d", s.port)})
}
