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

	// Run a slave server
	slave, err := master.RunSlaveServer(6380)
	if err != nil {
		log.Fatal(err)
	}
	defer slave.Stop()

	// Set value to the master
	c := redis.NewClient(&redis.Options{Addr: master.Addr()})

	if err := c.Set("foo", "bar", 0).Err(); err != nil {
		log.Fatal(err)
	}

	c.Close()

	// Get value from the slave
	c = redis.NewClient(&redis.Options{Addr: slave.Addr()})

	val, err := c.Get("foo").Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("foo", val)
}
