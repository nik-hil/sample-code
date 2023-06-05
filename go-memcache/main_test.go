package main

import (
	"net"
	"testing"
	"time"
)

type mockConn struct {
	readBuffer  []byte
	writeBuffer []byte
}

func (c *mockConn) Read(b []byte) (int, error) {
	n := copy(b, c.readBuffer)
	c.readBuffer = c.readBuffer[n:]
	return n, nil
}

func (c *mockConn) Write(b []byte) (int, error) {
	c.writeBuffer = append(c.writeBuffer, b...)
	return len(b), nil
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) LocalAddr() net.Addr {
	return nil
}

func (c *mockConn) RemoteAddr() net.Addr {
	return nil
}

func (c *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestCache_Set(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	cache.Set("key1", "value1", 10*time.Second)

	expected := "value1"
	actual, found := cache.Get("key1")

	if !found {
		t.Errorf("Expected key1 to be in cache but it was not found")
	}

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}

func TestCache_Get_ExpiredKey(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	cache.Set("key1", "value1", -10*time.Second)

	expected := ""
	actual, found := cache.Get("key1")

	if found {
		t.Errorf("Expected key1 to not be in cache but it was found")
	}

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}

func TestCache_Get_ValidKey(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	cache.Set("key1", "value1", 10*time.Second)

	expected := "value1"
	actual, found := cache.Get("key1")

	if !found {
		t.Errorf("Expected key1 to be in cache but it was not found")
	}

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}

func TestHandleConnection_InvalidCommand(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	conn := &mockConn{
		readBuffer:  []byte("invalid command\r\n"),
		writeBuffer: make([]byte, 0),
	}

	handleConnection(conn, cache)

	expected := "ERROR\r\n"
	actual := string(conn.writeBuffer)

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}

func TestHandleConnection_InvalidSetCommand(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	conn := &mockConn{
		readBuffer:  []byte("set key1\r\n"),
		writeBuffer: make([]byte, 0),
	}

	handleConnection(conn, cache)

	expected := "ERROR\r\n"
	actual := string(conn.writeBuffer)

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to not be in cache but it was found")
	}
}

func TestHandleConnection_SetCommand(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	conn := &mockConn{
		readBuffer:  []byte("set key1 value1 10\r\n"),
		writeBuffer: make([]byte, 0),
	}

	handleConnection(conn, cache)

	expected := "STORED\r\n"
	actual := string(conn.writeBuffer)

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}

	_, found := cache.Get("key1")
	if !found {
		t.Errorf("Expected key1 to be in cache but it was not found")
	}
}

func TestHandleConnection_GetCommand(t *testing.T) {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	cache.Set("key1", "value1", 100*time.Second)
	cache.Set("key2", "value2", 200*time.Second)

	conn := &mockConn{
		readBuffer:  []byte("get key1 key2\r\n"),
		writeBuffer: make([]byte, 0),
	}

	go handleConnection(conn, cache)

	expected := "VALUE key1 6\r\nvalue1\r\nVALUE key2 6\r\nvalue2\r\nEND\r\n"
	actual := string(conn.writeBuffer)

	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}
