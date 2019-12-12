package onkyoctl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ConnectionState is the type used to describe the connection status for the client.
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
	Disconnecting
)

var (
	ErrNotConnected = errors.New("not connected")
    ErrTimeout      = errors.New("timeout")
)

// MessageHandler is a callback function to handle incoming messages.
type MessageHandler func(ISCPCommand)

type sendTask struct {
    Command ISCPCommand
    Reply   chan error
}

type client struct {
	host           string
	port           int
	timeout        time.Duration
	state          ConnectionState
	conn           net.Conn
	connLock       sync.Mutex
	done           chan bool
	wantConnect    chan bool
	wantDisconnect chan bool
	received       chan ISCPCommand
	send           chan sendTask
	handler        MessageHandler
	connectionCB   func(ConnectionState)
	log            Logger
}

func newClient(host string, port int, log Logger) *client {
	return &client{
		host:           host,
		port:           port,
		timeout:        3 * time.Second,
		state:          Disconnected,
		done:           make(chan bool),
		wantConnect:    make(chan bool),
		wantDisconnect: make(chan bool),
		received:       make(chan ISCPCommand, 32),
		send:           make(chan sendTask, 32),
		log:            log,
	}
}

// public interface -----------------------------------------------------------

func (c *client) Start() {
	// if started, ignore
	go c.loop()
}

func (c *client) Stop() {
	// if stopped, ignore
	c.done <- true
}

func (c *client) Connect() {
	c.wantConnect <- true
}

func (c *client) Disconnect() {
	c.wantDisconnect <- true
}

func (c *client) State() ConnectionState {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	return c.state
}

func (c *client) Send(cmd ISCPCommand, timeout time.Duration) error {
	if c.isState(Disconnected, Disconnecting) {
		return ErrNotConnected
	}
    reply := make(chan error, 1)
	c.send <- sendTask{Command: cmd, Reply: reply}

    if timeout <= 0 {
        return nil
    }

    select{
    case err := <-reply:
        return err
    case <-time.After(timeout):
        return ErrTimeout
    }
}

func (c *client) loop() {
	for {
		select {
		case <-c.done:
			c.doDone()
			return
		case <-c.wantDisconnect:
			c.doDisconnect()
		case <-c.wantConnect:
			c.doConnect()
		case cmd := <-c.received:
			c.doReceive(cmd)
		case task := <-c.send:
			c.doSend(task)
		}
	}
}

func (c *client) doDone() {
	c.log.Debug("Done")
	c.doDisconnect()
}

// Connection handling --------------------------------------------------------

func (c *client) isState(states ...ConnectionState) bool {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	for _, s := range states {
		if s == c.state {
			return true
		}
	}
	return false
}

func (c *client) changeState(s ConnectionState, conn net.Conn) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.state = s
	if conn != nil {
		c.conn = conn
	}

	if c.connectionCB != nil {
		go func() {
			c.connectionCB(s)
		}()
	}
}

func (c *client) doConnect() {
	if c.isState(Connected, Connecting) {
		return
	}
	c.log.Debug("Connect")

	c.changeState(Connecting, nil)

	conn, err := c.createConn()
	if err != nil {
		c.changeState(Disconnected, nil)
		return
	}

	// we are connected
	c.changeState(Connected, conn)
	go c.readLoop(c.conn) // TODO: not thread safe
}

func (c *client) createConn() (net.Conn, error) {
	addr := fmt.Sprintf("%v:%v", c.host, c.port)
	return net.DialTimeout(protocol, addr, c.timeout)
}

func (c *client) doDisconnect() {
	if c.isState(Disconnected, Disconnecting) {
		return
	}
	c.log.Debug("Disconnect")

	c.changeState(Disconnecting, c.conn)
	// wait for outgoing messages?
	err := c.conn.Close() // TODO: not thread safe
	if err != nil {
		c.log.Warning("Error closing connection: %v", err)
	}
	c.changeState(Disconnected, nil)
}

func (c *client) xreconnect() {
	c.log.Debug("Schedule reconnect")
	go func() {
		time.Sleep(5 + time.Second)
		c.Connect()
	}()
}

func (c *client) readLoop(conn net.Conn) {
	defer func() {
		if c.isState(Connected) {
			// unexpected close of connection, assume server side close
			// and attempt reconnect
			c.changeState(Disconnected, nil)
		}
	}()

	r := bufio.NewReader(conn)
	buf := make([]byte, headerSize) // reused

	for {
		// read header
		_, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				// assume server side close
				return
			}
			c.log.Warning("Read error: %v", err)
			// return
			continue
		}
		c.log.Debug("<- recv (H): %v", buf)
		_, payloadSize, err := ParseHeader(buf)
		if err != nil {
			c.log.Warning("Discard bad message: %v", err)
			continue
		}

		// read payload
		payload := make([]byte, payloadSize)
		_, err = r.Read(payload)
		if err != nil {
			if err == io.EOF {
				// assume server side close
				return
			}
			c.log.Warning("Read error: %v", err)
			//return
			continue
		}
		c.log.Debug("<- recv (P): %v", payload)

		iscp, err := ParseISCP(payload)
		if err != nil {
			c.log.Warning("Discard invalid message: %v", err)
			continue
		}

		c.received <- iscp.Command()
	}
}

// send + receive -------------------------------------------------------------

func (c *client) doSend(t sendTask) {
	if !c.isState(Connected) {
		c.log.Warning("Discard message (not connected): %v", t.Command)
        t.Reply <- ErrNotConnected
		return
	}
	conn := c.conn // TODO: not thread safe

	msg := NewEISCPMessage(t.Command)
	c.log.Debug("-> send: %v", t.Command)
	_, err := conn.Write(msg.Raw())
	if err != nil {
		c.log.Error("Error writing to connection: %v", err)
	}
    t.Reply <- err
}

func (c *client) doReceive(cmd ISCPCommand) {
	c.log.Debug("<- handle: %v", cmd)
	if c.handler != nil {
		c.handler(cmd)
	}
}

// pretty print for connection state

func (cs ConnectionState) String() string {
	switch cs {
	case Connected:
		return "CONNECTED"
	case Connecting:
		return "CONNECTING"
	case Disconnected:
		return "DISCONNECTED"
	case Disconnecting:
		return "DISCONNECTING"
	default:
		return fmt.Sprintf("UNKNOWN (%v)", int(cs))
	}
}
