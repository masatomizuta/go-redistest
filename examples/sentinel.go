package main

import (
	"fmt"
	"log"

	"gopkg.in/redis.v5"

	"github.com/masatomizuta/go-redistest"
)

func main() {
	// Run a master server
	master, err := redistest.RunServer(6379)
	if err != nil {
		log.Fatal(err)
	}
	defer master.Stop()

	// Run a sentinel server
	sentinel, err := master.RunSentinelServer(26379, "mymaster")
	if err != nil {
		log.Fatal(err)
	}
	defer sentinel.Stop()

	// Access through sentinel
	c := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs: []string{sentinel.Addr()},
		MasterName: "mymaster",
	})

	if err := c.Set("foo", "bar", 0).Err(); err != nil {
		log.Fatal(err)
	}

	val, err := c.Get("foo").Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("foo", val)
}
