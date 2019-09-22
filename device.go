package onkyoctl

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	defaultPort = 60128
	protocol    = "tcp"
)

// Device is an Onkyo device
type Device struct {
	Host     string
	Port     int
	commands CommandSet
	timeout  int
	conn     net.Conn
	send     chan ISCPCommand
	recv     chan ISCPCommand
}

// NewDevice sets up a new Onkyo device.
func NewDevice(host string) Device {
	return Device{
		Host:     host,
		Port:     defaultPort,
		commands: basicCommands(),
		timeout:  10,
		send:     make(chan ISCPCommand, 16),
		recv:     make(chan ISCPCommand, 16),
	}
}

// Start connects to the device and starts receiving messages.
func (d *Device) Start() error {
	log.Printf("Start %v", d)
	// TODO: if already started return err
	err := d.connect()
	if err != nil {
		return err
	}
	go d.loop()
	return nil
}

// Stop disconnects from the device and stop message processing.
func (d *Device) Stop() {
	log.Printf("Stop %v", d)
	// if not started, return

	close(d.recv)
	close(d.send)
	d.disconnect()
}

// SendCommand sends an "friendly" command (e.g. "power off") to the device.
func (d *Device) SendCommand(name string, param interface{}) error {
	log.Printf("Send command %v: %v", name, param)
	command, err := d.commands.CreateCommand(name, param)
	if err != nil {
		return err
	}
	d.send <- command
	return nil
}

// SendISCP sends a raw ISCP command to the device.
func (d *Device) SendISCP(command ISCPCommand) {
	d.send <- command
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
	msg := NewEISCPMessage(command)
	log.Printf("Send message: %v", msg)
	numWritten, err := d.conn.Write(msg.Raw())
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
	} else {
		log.Printf("Wrote %v bytes", numWritten)
	}
}

func (d *Device) doReceive(command ISCPCommand) {
	log.Printf("Recv message: %v", command)
	name, value, err := d.commands.ReadCommand(command)
	if err != nil {
		log.Printf("Error reading ISCP %q: %v", command, err)
		return
	}
	log.Printf("Received '%v %v'", name, value)
	// TODO: callback
}

func (d *Device) connect() error {
	addr := fmt.Sprintf("%v:%v", d.Host, d.Port)
	log.Printf("Connect to %v", addr)
	timeout := time.Duration(d.timeout) * time.Second
	conn, err := net.DialTimeout(protocol, addr, timeout)
	if err != nil {
		return err
	}
	// TODO: maybe handshake to check we are connected to an onkyo device?
	log.Println("Connected.")
	d.conn = conn
	go d.read()
	return nil
}

func (d *Device) disconnect() {
	if d.conn == nil {
		// not connected
		return
	}
	err := d.conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}

func (d *Device) read() {
	for {
		// read message header
		buf := make([]byte, headerSize)
		numRead, err := d.conn.Read(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			// host closes (EOF) when another client connects?
			return
		}
		log.Printf("Read: %v - %v", numRead, buf)
		_, payloadSize, err := ParseHeader(buf)
		if err != nil {
			log.Printf("bad message data: %v", err)
			continue
		}

		// read message payload
		payload := make([]byte, payloadSize)
		numPayload, err := d.conn.Read(payload)
		if err != nil {
			log.Printf("Read error: %v", err)
			// host closes (EOF) when another client connects?
			return
		}
		log.Printf("Read %v payload bytes", numPayload)

		iscp, err := ParseISCP(payload)
		if err != nil {
			log.Printf("invalid ISCP message: %v", err)
			continue
		}
		d.recv <- iscp.Command()

	}
}

func basicCommands() CommandSet {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
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
			Group:     "SPA",
			ParamType: "onOff",
		},
	}
	return NewBasicCommandSet(commands)
}
