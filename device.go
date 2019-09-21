package onkyoctl

import (
    "fmt"
    "log"
    "net"
)

const (
    defaultPort = 60128
)

type Device struct {
    Host string
    Port int
    conn net.Conn
    send chan *EISCPMessage
    recv chan *EISCPMessage
}

func NewDevice(host string) Device {
    return Device{
        Host: host,
        Port: defaultPort,
        send: make(chan *EISCPMessage, 16),
        recv: make(chan *EISCPMessage, 16),
    }
}

func (d *Device) Start() error {
    // TODO: if already started return err
    err := d.connect()
    if err != nil {
        return err
    }
    go d.loop()
    return nil
}

func (d *Device) Stop() {
    // if not started, return

    close(d.recv)
    close(d.send)
    err := d.conn.Close()
    if err != nil {
        log.Printf("Error closing connection: %v", err)
    }
}

func (d *Device) SendCommand(command ISCPCommand) {
    msg := NewEISCPMessage(command)
    d.send <- msg
}

func (d *Device) loop() {
    for {
        select {
        case msg, more := <-d.recv:
            if more {
                d.doReceive(msg)
            }
        case msg, more := <-d.send:
            if more {
                d.doSend(msg)
            }
        }
    }
}

func (d *Device) doSend(msg *EISCPMessage) {
    log.Printf("Send message: %v", msg)
    numWritten, err := d.conn.Write(msg.Raw())
    if err != nil {
        log.Printf("Error writing to connection: %v", err)
    } else {
        log.Printf("Wrote %v bytes", numWritten)
    }
}

func (d *Device) doReceive(msg *EISCPMessage) {
    log.Printf("Recv message: %v", msg)
    // TODO: callback
}

func (d *Device) connect() error {
    addr := fmt.Sprintf("%v:%v", d.Host, d.Port)
    log.Printf("Connect to %v", addr)
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return err
    }
    // TODO: maybe handshake to check we are connected to an onkyo device?
    d.conn = conn
    go d.read()
    return nil
}

func (d *Device) read() {
    data := make([]byte, 0)
    for {
        // read up to N bytes ...
        buf := make([]byte, 32)
        numRead, err := d.conn.Read(buf)
        if err != nil {
            log.Printf("Read error: %v", err)
            return
        }

        // ... combine with what we already have
        data = append(data, buf[:numRead]...)

        //TODO: make a loop until all of data is parsed

        // parse "consumes" some of our data
        msg, err := ParseEISCP(data)
        consumed := 0
        if err != nil {
            // parse error
        } else {
            d.recv <- msg
            // TODO: the parser needs to tell us
            consumed = len(data)
        }

        // keep the remainder
        data = data[consumed:]
    }
}
