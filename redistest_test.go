package redistest

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/redis.v5"
)

func TestRunMasterServer(t *testing.T) {
	s, err := RunServer(DefaultMasterPort)
	defer s.Stop()
	assert.NoError(t, err)

	s2, err := RunServer(DefaultMasterPort)
	assert.Nil(t, s2)
	assert.Error(t, err)
}

func TestServer_RunAndStop(t *testing.T) {
	s, err := RunServer(DefaultMasterPort)
	assert.NoError(t, err)

	err = s.Run()
	assert.Error(t, err)

	err = s.Stop()
	assert.NoError(t, err)

	err = s.Stop()
	assert.Error(t, err)

	err = s.Run()
	assert.NoError(t, err)

	err = s.Stop()
	assert.NoError(t, err)
}

func TestServer_RunSlaveServer(t *testing.T) {
	m, err := RunServer(DefaultMasterPort)
	defer m.Stop()
	assert.NoError(t, err)

	s, err := m.RunSlaveServer(DefaultSlavePort)
	defer s.Stop()
	assert.NoError(t, err)
	assert.True(t, s.IsRunning())

	c := redis.NewClient(&redis.Options{Addr: fmt.Sprintf(":%d", DefaultMasterPort)})
	err = c.Set("foo", "bar", 0).Err()
	assert.NoError(t, err)
	c.Close()

	// Wait for the replication sync
	time.Sleep(500 * time.Millisecond)

	c = redis.NewClient(&redis.Options{Addr: fmt.Sprintf(":%d", DefaultSlavePort)})
	bar, err := c.Get("foo").Result()
	assert.NoError(t, err)
	assert.Equal(t, "bar", bar)
	c.Close()
}

func TestServer_RunSentinelServer(t *testing.T) {
	m, err := RunServer(DefaultMasterPort)
	defer m.Stop()
	assert.NoError(t, err)

	s, err := m.RunSentinelServer(DefaultSentinelPort, DefaultSentinelMaster)
	defer s.Stop()
	assert.NoError(t, err)
	assert.True(t, s.IsRunning())

	c := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    DefaultSentinelMaster,
		SentinelAddrs: []string{fmt.Sprintf("localhost:%d", DefaultSentinelPort)},
	})
	err = c.Set("foo", "bar", 0).Err()
	assert.NoError(t, err)
	c.Close()

	c = redis.NewClient(&redis.Options{Addr: fmt.Sprintf(":%d", DefaultMasterPort)})
	bar, err := c.Get("foo").Result()
	assert.NoError(t, err)
	assert.Equal(t, "bar", bar)
	c.Close()
}

func TestServer_IsRunning(t *testing.T) {
	s, err := RunServer(DefaultMasterPort)
	assert.NoError(t, err)
	assert.True(t, s.IsRunning())

	err = s.Stop()
	assert.NoError(t, err)
	assert.False(t, s.IsRunning())
}

func TestServer_Port(t *testing.T) {
	s, err := RunServer(DefaultMasterPort)
	defer s.Stop()

	assert.NoError(t, err)
	assert.Equal(t, DefaultMasterPort, s.Port())
}

func TestServer_Flush(t *testing.T) {
	s, err := RunServer(DefaultMasterPort)
	defer s.Stop()

	c := s.NewClient()
	defer c.Close()

	c.Set("foo", "bar", 0)

	n, err := c.DbSize().Result()
	assert.NoError(t, err)
	assert.Equal(t, 1, int(n))

	err = s.Flush()
	assert.NoError(t, err)

	n, err = c.DbSize().Result()
	assert.NoError(t, err)
	assert.Zero(t, n)
}
