package onkyoctl

import (
	"fmt"
	"log"
	"net"
)

const (
	defaultPort = 60128
	protocol    = "tcp"
)

// Device is an Onkyo device
type Device struct {
	Host string
	Port int
	conn net.Conn
	send chan ISCPCommand
	recv chan ISCPCommand
}

// NewDevice sets up a new Onkyo device.
func NewDevice(host string) Device {
	return Device{
		Host: host,
		Port: defaultPort,
		send: make(chan ISCPCommand, 16),
		recv: make(chan ISCPCommand, 16),
	}
}

// Start connects to the device and strts receiving messages.
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

// SendCommand sends an ISCP command to the device.
func (d *Device) SendCommand(command ISCPCommand) {
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
	// TODO: callback
}

func (d *Device) connect() error {
	addr := fmt.Sprintf("%v:%v", d.Host, d.Port)
	log.Printf("Connect to %v", addr)
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		return err
	}
	// TODO: Timeouts
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
