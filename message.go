package onkyoctl

import (
	"encoding/binary"
	"errors"
)

const (
	iscpStart               = "!"
	unitTypeReceiver        = "1"
	headerSize       uint32 = 16
	eISCPVersion     byte   = 0x01
	terminator              = "\r\n"
	cr                      = byte('\r')
	lf                      = byte('\n')
	eof                     = 0x1A
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
	return iscpStart + unitTypeReceiver + string(i.command) + terminator
}

// Command returns the ISCP command for this message.
func (i *ISCPMessage) Command() ISCPCommand {
	return i.command
}

func (i *ISCPMessage) String() string {
	return "ISCP " + string(i.Command())
}

// ToEISCP converts this message to eISCP format.
func (i *ISCPMessage) ToEISCP() *EISCPMessage {
	return &EISCPMessage{
		message: i,
	}
}

// EISCPMessage is the type for eISCP messages.
type EISCPMessage struct {
	message *ISCPMessage
}

// NewEISCPMessage creates a new eISCP message for the given command.
func NewEISCPMessage(command ISCPCommand) *EISCPMessage {
	return NewISCPMessage(command).ToEISCP()
}

// Command returns the ISCP command for this message.
func (e *EISCPMessage) Command() ISCPCommand {
	return e.message.Command()
}

func (e *EISCPMessage) String() string {
	return "eISCP " + string(e.Command())
}

// Raw returns the byte data (header and payload) for this message.
func (e *EISCPMessage) Raw() []byte {
	payload := []byte(e.message.Format())

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
	header[0] = 0x49 // I
	header[1] = 0x53 // S
	header[2] = 0x43 // C
	header[3] = 0x50 // P
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
	return result
}

// ParseEISCP reads an eISCP message from a byte array.
func ParseEISCP(data []byte) (*EISCPMessage, error) {
	headerSize, payloadSize, err := ParseHeader(data)
	if err != nil {
		return nil, err
	}

	totalSize := headerSize + payloadSize
	if len(data) < totalSize {
		return nil, errors.New("eISCP message too short")
	}

	payload := data[headerSize:totalSize]
	iscp, err := ParseISCP(payload)
	if err != nil {
		return nil, err
	}
	return iscp.ToEISCP(), nil
}

// ParseHeader parses the header of an eISCP message
// and returns the header size and payload size
func ParseHeader(data []byte) (int, int, error) {
	// we need at least 12 byte
	// - 4 bytes "magic"
	// - 4 bytes header length
	// - 4 bytes payload length
	if len(data) < 12 {
		return 0, 0, errors.New("eISCP header too short")
	}

	// check the "magic"
	iOk := data[0] == 0x49 // I
	sOk := data[1] == 0x53 // S
	cOk := data[2] == 0x43 // C
	pOk := data[3] == 0x50 // P
	magicOk := iOk && sOk && cOk && pOk
	if !magicOk {
		return 0, 0, errors.New("missing start sequence in message header")
	}

	end := binary.BigEndian
	headerSize := end.Uint32(data[4:8])
	payloadSize := end.Uint32(data[8:12])
	if len(data) < int(headerSize) {
		return 0, 0, errors.New("eISCP header shorter than indicated")
	}

	// note we might parse the last 4 bytes to get the version

	return int(headerSize), int(payloadSize), nil
}

// ParseISCP parses an ISCP message from a byte array.
func ParseISCP(data []byte) (*ISCPMessage, error) {
	// decode to string first
	s := string(data)
	size := len(s)

	// expect: !1<COMMAND>\r\n
	// where Command is at least three digits
	// we can do without CR/LF at the end
	if size < 5 {
		return nil, errors.New("ISCP message too short")
	}
	if s[0] != byte('!') {
		return nil, errors.New("missing start character '!'")
	}
	if s[1] != byte('1') {
		return nil, errors.New("missing receiver type '1'")
	}

	// terminators can be:
	// - LF     1 byte
	// - CR     1 byte
	// - CRLF   2 bytes
	// - <none>
	offset := size - 1
	if s[offset] == cr { // CR
		offset--

	} else if s[offset] == lf { // LF or CRLF
		offset--
		if s[offset] == cr { // CRLF
			offset--
		}
	}

	// for messages from device to controller, EOF should be mandatory.
	// however, we don't mind if it is missing.
	if s[offset] == eof {
		offset--
	}

	command := string(s[2 : offset+1])
	return NewISCPMessage(ISCPCommand(command)), nil
}
