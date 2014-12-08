package main

import (
	"github.com/fzzy/radix/redis"
	"sync"
	"time"
)

type RedisPool struct {
	// Max number of connections to keep
	MaxConnections int
	// Address to connect to
	Address string
	// Auth password to use, if any
	Auth string
	// List of connections
	Clients []*redis.Client
	// The "place" we are in the conneciton list, for round-robining
	Offset int
}

// Creates a new redis pool
func CreateRedisPool(address string, connections int, auth string) *RedisPool {
	pool := &RedisPool{
		Address:        address,
		MaxConnections: connections,
		Auth:           auth,
		Clients:        []*redis.Client{},
	}

	var wg sync.WaitGroup
	wg.Add(connections)
	// Create the number of connections we want asynchronously.
	for n := 0; n < connections; n += 1 {
		go func() {
			pool.CreateClient()
			defer wg.Done()
		}()
	}

	wg.Wait()
	go pool.Monitor()

	return pool
}

// Pings open connections to make sure they're active
func (p *RedisPool) Monitor() {
	for _ = range time.Tick(5 * time.Second) {
		for _, client := range p.Clients {
			output, _ := client.Cmd("PING").Str()

			if output != "PONG" {
				log.Info("Server down, todo reconnect")
			}
		}
	}
}

// Creates a new connection and adds it on the pool.
func (p *RedisPool) CreateClient() error {
	print("creating connection\n")
	client, err := redis.Dial("tcp", p.Address)
	if err == nil {
		p.Clients = append(p.Clients, client)

		// If we set an auth password, send the auth command.
		if len(p.Auth) > 0 {
			client.Cmd("AUTH", config.Redis.Auth)
		}
	} else {
		log.Error("Redis error: %s", err)
		client.Close()
	}

	return err
}

// Retrieves a connection from the pool
func (p *RedisPool) Get() *redis.Client {
	// Update the round-robin counter
	if p.Offset >= len(p.Clients) {
		p.Offset = 0
	}
	client := p.Clients[p.Offset]
	p.Offset += 1
	return client
}
