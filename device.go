package onkyoctl

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const protocol = "tcp"

// Callback is the type for message callback functions.
type Callback func(name, value string)

// Device is an Onkyo device
type Device struct {
	Host           string
	Port           int
	log            Logger
	commands       CommandSet
	callback       Callback
	onConnect      func()
	onDisconnect   func()
	timeout        int
	conn           net.Conn
	send           chan ISCPCommand
	recv           chan ISCPCommand
	wait           *sync.WaitGroup
	reco           *time.Timer
	autoConnect    bool
	allowReconnect bool
	reconnectTime  time.Duration
	isRunning      bool
}

// NewDevice sets up a new Onkyo device.
func NewDevice(cfg *Config) *Device {
	commands := cfg.Commands
	if commands == nil {
		commands = emptyCommands()
	}

	log := cfg.Log
	if log == nil {
		log = NewLogger(NoLog)
	}

	return &Device{
		Host:           cfg.Host,
		Port:           cfg.Port,
		log:            log,
		commands:       commands,
		timeout:        cfg.ConnectTimeout,
		wait:           &sync.WaitGroup{},
		send:           make(chan ISCPCommand, 16),
		recv:           make(chan ISCPCommand, 16),
		autoConnect:    cfg.AutoConnect,
		allowReconnect: cfg.AllowReconnect,
		reconnectTime:  time.Duration(cfg.ReconnectSeconds) * time.Second,
	}
}

// OnMessage sets the handler for received messages to the given function.
// This will replace any existing handler.
func (d *Device) OnMessage(callback Callback) {
	d.callback = callback
}

// OnDisconnected is called when the device is disconnected.
func (d *Device) OnDisconnected(callback func()) {
	d.onDisconnect = callback
}

// OnConnected is called when the deivce is (re-)connected.
func (d *Device) OnConnected(callback func()) {
	d.onConnect = callback
}

// Start connects to the device and starts receiving messages.
func (d *Device) Start() error {
	if d.isRunning {
		return errors.New("already started")
	}
	d.log.Info("Start device [%v:%v]", d.Host, d.Port)

	d.reco = time.NewTimer(d.reconnectTime)

	err := d.connect()
	if err != nil {
		return err
	}

	d.isRunning = true
	go d.loop()

	return nil
}

// Stop disconnects from the device and stop message processing.
func (d *Device) Stop() {
	d.log.Info("Stop device [%v:%v]", d.Host, d.Port)
	if !d.isRunning {
		return
	}

	d.disconnect()
	close(d.recv)
	close(d.send)
}

// SendCommand sends an "friendly" command (e.g. "power off") to the device.
//
// This method calls `SendISCP()` behind the scenes.
func (d *Device) SendCommand(name string, param interface{}) error {
	command, err := d.commands.CreateCommand(name, param)
	if err != nil {
		return err
	}

	return d.SendISCP(command)
}

// Query sends a QSTN command for the given friendly name.
//
// This method calls `SendISCP()` behind the scenes.
func (d *Device) Query(name string) error {
	q, err := d.commands.CreateQuery(name)
	if err != nil {
		return err
	}
	return d.SendISCP(q)
}

// SendISCP sends a raw ISCP command to the device.
//
// You must `Start()` before you can send messages.
// The device may lose its connection after start. With AutoConnect
// set to true, attempts to connect and returns an error only if that fails.
// Without autoconnect, an error is returned if the device is not connected.
//
// The message is send asynchronously. Use `WaitSend()` to block until the
// message is actually transmitted.
func (d *Device) SendISCP(command ISCPCommand) error {
	if !d.isRunning {
		return errors.New("device not started")
	}

	err := d.connectIfRequired()
	if err != nil {
		return err
	}

	d.log.Debug("Dispatch %v", command)

	d.wait.Add(1)
	d.send <- command
	return nil
}

// WaitSend waits for all submitted messages to be sent.
func (d *Device) WaitSend(timeout time.Duration) {
	done := make(chan int)
	go func() {
		defer close(done)
		d.wait.Wait()
	}()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		return
	}
}

func (d *Device) loop() {
	for {
		select {
		case <-d.reco.C:
			d.reconnect()
		case command, more := <-d.recv:
			if more {
				d.doReceive(command)
			}
		case command, more := <-d.send:
			if more {
				d.doSend(command)
			}
		}
	}
}

func (d *Device) doSend(command ISCPCommand) {
	defer d.wait.Done()

	msg := NewEISCPMessage(command)
	d.log.Debug("Send (TCP): %v", msg)
	_, err := d.conn.Write(msg.Raw())
	if err != nil {
		d.log.Error("Error writing to connection: %v", err)
	}
}

func (d *Device) doReceive(command ISCPCommand) {
	d.log.Debug("Receive message: %v", command)

	name, value, err := d.commands.ReadCommand(command)
	if err != nil {
		d.log.Warning("Error reading %q: %v", command, err)
		return
	}
	d.log.Debug("Received '%v %v'", name, value)
	if d.callback != nil {
		d.callback(name, value)
	}
}

func (d *Device) connect() error {
	d.log.Info("Connect to %v:%v ...", d.Host, d.Port)

	d.reco.Stop()

	addr := fmt.Sprintf("%v:%v", d.Host, d.Port)
	timeout := time.Duration(d.timeout) * time.Second
	conn, err := net.DialTimeout(protocol, addr, timeout)
	if err != nil {
		return err
	}

	d.log.Info("Connected.")
	d.conn = conn
	go d.read()

	// TODO: maybe handshake to check we are connected to an onkyo device?

	if d.onConnect != nil {
		go d.onConnect()
	}
	return nil
}

func (d *Device) isConnected() bool {
	return d.conn != nil
}

func (d *Device) connectIfRequired() error {
	if d.isConnected() {
		return nil
	}

	if d.autoConnect {
		return d.connect()
	}

	return errors.New("device not connected")
}

func (d *Device) disconnect() {
	if !d.isConnected() {
		return
	}
	d.log.Debug("Disconnect.")
	err := d.conn.Close()
	if err != nil {
		d.log.Warning("Error closing connection: %v", err)
	}
	d.conn = nil

	if d.onDisconnect != nil {
		go d.onDisconnect()
	}
}

func (d *Device) connectionLost() {
	// host closes connection when another client connects
	d.log.Error("Connection closed by remote device.")
	d.disconnect()
	if d.allowReconnect {
		d.reco.Reset(d.reconnectTime)
	}
}

func (d *Device) reconnect() {
	d.log.Debug("Reconnect...")
	if d.isConnected() {
		return
	}
	err := d.connect()
	if err != nil {
		d.log.Error("Reconnect failed: %v", err)
		// schedule the next attempt
		if d.allowReconnect {
			d.reco.Reset(d.reconnectTime)
		}
	}
}

func (d *Device) read() {
	for {
		// TODO: not thread-safe
		if !d.isConnected() {
			return
		}

		// read message header
		buf := make([]byte, headerSize)
		numRead, err := d.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				d.connectionLost()
				return
			}
			d.log.Warning("Read error: %v", err)
			return
		}
		d.log.Debug("Read header (%v): %v", numRead, buf)
		_, payloadSize, err := ParseHeader(buf)
		if err != nil {
			d.log.Warning("Discard bad message: %v", err)
			continue
		}

		// read message payload
		payload := make([]byte, payloadSize)
		numPayload, err := d.conn.Read(payload)
		if err != nil {
			if err == io.EOF {
				d.connectionLost()
				return
			}
			d.log.Warning("Read error: %v", err)
			return
		}
		d.log.Debug("Read payload (%v): %v", numPayload, payload)

		iscp, err := ParseISCP(payload)
		if err != nil {
			d.log.Warning("Discard invalid message: %v", err)
			continue
		}
		d.recv <- iscp.Command()
	}
}

// BasicCommands creates a command set with some commonly used commands.
func BasicCommands() CommandSet {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
		},
		Command{
			Name:      "volume",
			Group:     "MVL",
			ParamType: "intRangeEnum",
			Lower:     0,
			Upper:     100,
			Scale:     2,
			Lookup: map[string]string{
				"UP":   "up",
				"DOWN": "down",
			},
		},
		Command{
			Name:      "mute",
			Group:     "AMT",
			ParamType: "onOffToggle",
		},
		Command{
			Name:      "speaker-a",
			Group:     "SPA",
			ParamType: "onOff",
		},
		Command{
			Name:      "speaker-b",
			Group:     "SPB",
			ParamType: "onOff",
		},
		Command{
			Name:      "dimmer",
			Group:     "DIM",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "bright",
				"01": "dim",
				"02": "dark",
				"03": "off",
				"08": "led-off",
			},
		},
		Command{
			Name:      "display",
			Group:     "DIF",
			ParamType: "enumToggle",
			Lookup: map[string]string{
				"00": "default",
				"01": "listening",
				"02": "source",
				"03": "mode-4",
			},
		},
		Command{
			Name:      "input",
			Group:     "SLI",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "video-1",
				"01": "cbl-sat",
				"02": "game",
				"03": "aux1",
				"20": "tv-tape",
			},
		},
		Command{
			Name:      "listen-mode",
			Group:     "LMD",
			ParamType: "enum",
			Lookup: map[string]string{
				"00":     "stereo",
				"STEREO": "stereo",
				"01":     "direct",
				"11":     "pure",
			},
		},
		Command{
			Name:      "update",
			Group:     "UPD",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "no-new-firmware",
				"01": "new-firmware",
			},
		},
	}
	return NewBasicCommandSet(commands)
}

func emptyCommands() CommandSet {
	return NewBasicCommandSet(make([]Command, 0))
}
