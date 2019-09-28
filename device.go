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
	Host      string
	Port      int
	commands  CommandSet
	callback  Callback
	timeout   int
	conn      net.Conn
	send      chan ISCPCommand
	recv      chan ISCPCommand
	wait      *sync.WaitGroup
	isRunning bool
}

// NewDevice sets up a new Onkyo device.
func NewDevice(cfg *Config) Device {
	return Device{
		Host:     cfg.Host,
		Port:     cfg.Port,
		commands: cfg.Commands,
		timeout:  cfg.ConnectTimeout,
		wait:     &sync.WaitGroup{},
		send:     make(chan ISCPCommand, 16),
		recv:     make(chan ISCPCommand, 16),
	}
}

// OnMessage sets the handler for received messages to the given function.
// This will replace any existing handler.
func (d *Device) OnMessage(cb Callback) {
	d.callback = cb
}

// Start connects to the device and starts receiving messages.
func (d *Device) Start() error {
	if d.isRunning {
		return errors.New("already started")
	}
	logInfo("Start device [%v:%v]", d.Host, d.Port)

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
	logInfo("Stop device [%v:%v]", d.Host, d.Port)
	if !d.isRunning {
		return
	}

	d.disconnect()
	close(d.recv)
	close(d.send)
}

// SendCommand sends an "friendly" command (e.g. "power off") to the device.
func (d *Device) SendCommand(name string, param interface{}) error {
	logDebug("Dispatch command %v: %v", name, param)

	if d.commands == nil {
		return errors.New("command set ist not defined")
	}

	command, err := d.commands.CreateCommand(name, param)
	if err != nil {
		return err
	}

	return d.SendISCP(command)
}

// Query sends a QSTN command for the given friendly name.
func (d *Device) Query(name string) error {
	if d.commands == nil {
		return errors.New("command set ist not defined")
	}

	q, err := d.commands.CreateQuery(name)
	if err != nil {
		return err
	}
	return d.SendISCP(q)
}

// SendISCP sends a raw ISCP command to the device.
func (d *Device) SendISCP(command ISCPCommand) error {
	if !d.isRunning {
		return errors.New("device not started")
	}

	logDebug("Dispatch %v", command)

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
	logDebug("Send (TCP): %v", msg)
	numWritten, err := d.conn.Write(msg.Raw())
	if err != nil {
		logError("Error writing to connection: %v", err)
	} else {
		logDebug("Wrote %v bytes", numWritten)
	}
}

func (d *Device) doReceive(command ISCPCommand) {
	logDebug("Receive message: %v", command)
	if d.commands == nil {
		logWarning("Command set ist not defined, ignoring ISCP message.")
		return
	}

	name, value, err := d.commands.ReadCommand(command)
	if err != nil {
		logWarning("Error reading %q: %v", command, err)
		return
	}
	logDebug("Received '%v %v'", name, value)
	if d.callback != nil {
		d.callback(name, value)
	}
}

func (d *Device) connect() error {
	logInfo("Connect to %v:%v ...", d.Host, d.Port)

	addr := fmt.Sprintf("%v:%v", d.Host, d.Port)
	timeout := time.Duration(d.timeout) * time.Second
	conn, err := net.DialTimeout(protocol, addr, timeout)
	if err != nil {
		return err
	}
	// TODO: maybe handshake to check we are connected to an onkyo device?
	logInfo("Connected.")
	d.conn = conn
	go d.read()
	return nil
}

func (d *Device) disconnect() {
	if d.conn == nil {
		// not connected
		return
	}
	logDebug("Disconnect.")
	err := d.conn.Close()
	if err != nil {
		logWarning("Error closing connection: %v", err)
	}
	d.conn = nil
}

func (d *Device) connectionClosed() {
	// TODO: apparently, host closes connection when another client connects
	// should we exit? or reconnect?
	logError("Connection closed by remote device.")
	d.Stop()
}

func (d *Device) read() {
	for {
		// read message header
		buf := make([]byte, headerSize)
		numRead, err := d.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				d.connectionClosed()
				return
			}
			logWarning("Read error: %v", err)
			return
		}
		logDebug("Read header (%v): %v", numRead, buf)
		_, payloadSize, err := ParseHeader(buf)
		if err != nil {
			logWarning("Discard bad message: %v", err)
			continue
		}

		// read message payload
		payload := make([]byte, payloadSize)
		numPayload, err := d.conn.Read(payload)
		if err != nil {
			if err == io.EOF {
				d.connectionClosed()
				return
			}
			logWarning("Read error: %v", err)
			return
		}
		logDebug("Read payload (%v): %v", numPayload, payload)

		iscp, err := ParseISCP(payload)
		if err != nil {
			logWarning("Discard invalid message: %v", err)
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
