package onkyoctl

import (
    "encoding/binary"
    "log"
)

// An ISCPCommand is a low-level command like PWR01 (power on)
// or MVLUP (master volume up).
type ISCPCommand string

const (
    iscpStart = "!"
    unitTypeReceiver = "1"
    headerSize uint32 = 16
    eISCPVersion byte = 0x01
    crlf = "\r\n"
)

// ISCPMessage is the base message for ISCP.
// The messages conststs of:
// !    - start character
// 1    - receiver type
// ...  - <command>
// \r\n - terminator
type ISCPMessage struct {
    command ISCPCommand
}


// NewISCPMessage creates a new ISCP message with the given command.
func NewISCPMessage(command ISCPCommand) *ISCPMessage {
    return &ISCPMessage{
        command: command,
    }
}

// Format returns this message as a string, including terminating newline.
func (i *ISCPMessage) Format() string {
    return iscpStart + unitTypeReceiver + string(i.command) + crlf
}

// ToEISCP converts this message to eISCP format.
func (i *ISCPMessage) ToEISCP() EISCPMessage {
    return EISCPMessage{
        message: i,
    }
}

// EISCPMessage is the type for eISCP messages.
type EISCPMessage struct {
    message *ISCPMessage
}

// NewEISCPMessage creates a new eISCP message for the given command.
func NewEISCPMessage(command ISCPCommand) EISCPMessage {
    return NewISCPMessage(command).ToEISCP()
}

// Raw returns the byte data (header and payload) for this message.
func (e *EISCPMessage) Raw() []byte {
    msg := e.message.Format()
    // TODO: encoding is utf-8
    payload := []byte(msg)

    end := binary.BigEndian

    headerLen := make([]byte, 4)
    end.PutUint32(headerLen, headerSize)

    msgLen := make([]byte, 4)
    end.PutUint32(msgLen, uint32(len(payload)))

    // Header
    // 0-3      magic 'ISCP'
    // 4-7      length of the header (always 16)
    // 8-11     length of the payload (in bytes)
    // 12       version
    // 13-15    reserved (0x00 0x00 0x00)
    header := make([]byte, headerSize)
    header[0] = 'I'
    header[1] = 'S'
    header[2] = 'C'
    header[3] = 'P'
    header[4] = headerLen[0]
    header[5] = headerLen[1]
    header[6] = headerLen[2]
    header[7] = headerLen[3]
    header[8] = msgLen[0]
    header[9] = msgLen[1]
    header[10] = msgLen[2]
    header[11] = msgLen[3]
    header[12] = eISCPVersion
    header[13] = 0x00
    header[14] = 0x00
    header[15] = 0x00

    result := append(header, payload...)
    log.Printf("Raw eISCP message: %v", result)
    return result
}
