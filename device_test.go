package onkyoctl

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

var validCommand = ISCPCommand("PWR01")
var testPort = 30128

func TestDeviceBasics(t *testing.T) {
	device := NewDevice(testConfig())

	err := device.SendISCP(validCommand, 0)
	if err == nil {
		t.Log("Missing expected error when using non-started device")
		t.Fail()
	}
}

func xTestDeviceConnectAndSend(t *testing.T) {
	device := NewDevice(testConfig())
	server := newMockServer()

	server.Start()
	defer server.Stop()

	device.Start()
	defer device.Stop()

	if !server.WaitConnected() {
		t.Log("Server does not see connection after device start")
		t.Fail()
	}

	device.SendISCP(validCommand, 0)

	data, err := server.ReadRaw()
	if err != nil {
		t.Log("error reading from device")
		t.Fail()
		return
	}

	msg, err := ParseEISCP(data)
	if err != nil {
		t.Logf("server received invalid message: %v", err)
		t.Fail()
		return
	}

	if msg.Command() != validCommand {
		t.Logf("server did not receive expected command")
		t.Fail()
	}
}

func xTestDeviceAutoConnect(t *testing.T) {
	cfg := testConfig()
	cfg.AutoConnect = true
	device := NewDevice(cfg)
	server := newMockServer()

	server.Start()
	defer server.Stop()

	device.Start()
	defer device.Stop()

	if !server.WaitConnected() {
		t.Log("initial connect failed")
		t.Fail()
		return
	}

	// check on the disconnected callback
	didCall := false
	device.OnDisconnected(func() {
		didCall = true
	})

	// close server-side and wait for client to realize new state
	server.Disconnect()
	time.Sleep(1 * time.Second)

	// check for the callback
	if !didCall {
		t.Log("OnDisconnect callback did not fire")
		t.Fail()
	}

	// device is started but disconnected
	// and should automatically reconnect as per config
	// we should see a new incoming connection server side and the command.
	//
	// TODO: does not work - device reconnects, but we do not see it
	err := device.SendISCP(validCommand, 0)
	if err != nil {
		t.Logf("unexpected send error: %v", err)
		t.Fail()
		return
	}

	if !server.WaitConnected() {
		t.Logf("device did not automatically re-connect")
		t.Fail()
		return
	}

	_, err = server.ReadRaw()
	if err != nil {
		t.Logf("Read error: %v", err)
		t.Fail()
	}
}

func testConfig() *Config {
	cfg := DefaultConfig()
	cfg.Port = testPort
	cfg.Host = "localhost"
	cfg.AutoConnect = false
	cfg.AllowReconnect = false
	cfg.ReconnectSeconds = 1

	//cfg.Log = NewLogger(NoLog)
	cfg.Log = NewLogger(Debug)
	return cfg
}

type mockServer struct {
	port      int
	connected chan (bool)
	data      chan ([]byte)
	listener  net.Listener
	conn      net.Conn
}

func newMockServer() *mockServer {
	return &mockServer{
		port:      testPort,
		connected: make(chan bool, 1),
		data:      make(chan []byte, 1),
	}
}

func (m *mockServer) Start() {
	addr := fmt.Sprintf("localhost:%v", m.port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	m.listener = l

	go m.listen()
}

func (m *mockServer) WaitConnected() bool {
	select {
	case result := <-m.connected:
		return result
	case <-time.After(200 * time.Millisecond):
		return false
	}
}

func (m *mockServer) listen() {
	for {
		conn, err := m.listener.Accept()

		// new connect closes existing
		m.Disconnect()

		if err != nil {
			m.connected <- false
			return
		}

		m.conn = conn
		m.connected <- true

		buf := make([]byte, 512) // buf size should be larger than expected message size
		numRead, err := conn.Read(buf)
		if err != nil {
			return
		}
		m.data <- buf[0:numRead]
	}
}

func (m *mockServer) ReadRaw() ([]byte, error) {
	select {
	case result := <-m.data:
		return result, nil
	case <-time.After(200 * time.Millisecond):
		return nil, errors.New("timeout")
	}
}

// Disconnect closes the client connection
func (m *mockServer) Disconnect() {
	if m.conn != nil {
		m.conn.Close()
	}
	m.conn = nil
	//m.connected = make(chan bool, 1)
	//m.data = make(chan []byte, 1)
}

func (m *mockServer) Stop() {
	m.Disconnect()

	if m.listener != nil {
		m.listener.Close()
	}
	m.listener = nil
}
