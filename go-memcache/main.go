package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type CacheItem struct {
	value      string
	expiration time.Time
}

type Cache struct {
	items map[string]CacheItem
}

func (c *Cache) Set(key string, value string, expiration time.Duration) {
	c.items[key] = CacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	item, found := c.items[key]
	if !found {
		return "", false
	}
	val := item.expiration.Compare(time.Now())
	if val < 0 {
		delete(c.items, key)
		return "", false
	}
	return item.value, true
}

func handleConnection(conn net.Conn, cache *Cache) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		command := strings.TrimSpace(string(buffer[:n]))
		parts := strings.Split(command, " ")
		if len(parts) < 2 {
			conn.Write([]byte("ERROR\r\n"))
			continue
		}

		switch parts[0] {
		case "set":
			if len(parts) < 4 {
				conn.Write([]byte("ERROR\r\n"))
				continue
			}
			key := parts[1]
			value := parts[2]
			expiration, err := strconv.Atoi(parts[3])
			if err != nil {
				conn.Write([]byte("ERROR\r\n"))
				continue
			}
			cache.Set(key, value, time.Duration(expiration)*time.Second)
			conn.Write([]byte("STORED\r\n"))
		case "get":
			for _, key := range parts[1:] {
				value, found := cache.Get(key)
				if found {
					conn.Write([]byte(fmt.Sprintf("VALUE %s %d\r\n%s\r\n", key, len(value), value)))
				}
			}
			conn.Write([]byte("END\r\n"))
		case "quit":
			return
		default:
			conn.Write([]byte("ERROR\r\n"))
		}
	}
}

func main() {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	listener, err := net.Listen("tcp", ":11211")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Listening on :11211")

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn, cache)
	}
}
