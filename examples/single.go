package main

import (
	"fmt"
	"log"

	"gopkg.in/redis.v5"

	"github.com/masatomizuta/go-redistest"
)

func main() {
	// Run a new redis server
	s, err := redistest.RunServer(6379)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Stop()

	// Access with client
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer c.Close()

	if err := c.Set("foo", "bar", 0).Err(); err != nil {
		log.Fatal(err)
	}

	val, err := c.Get("foo").Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("foo", val)
}
