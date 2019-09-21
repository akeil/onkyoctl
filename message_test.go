package onkyoctl

import (
    "encoding/binary"
	"testing"
)

func TestISCPFormat(t *testing.T) {
    var command ISCPCommand
    command = "PWR01"
    msg := NewISCPMessage(command)
    s := msg.Format()
    assertEqual(t, s, "!1PWR01\r\n")
}


func TestEISCPRaw(t *testing.T) {
    m := NewEISCPMessage("PWR01")
    raw := m.Raw()

    end := binary.BigEndian

    assertEqual(t, raw[0], byte('I'))
    assertEqual(t, raw[1], byte('S'))
    assertEqual(t, raw[2], byte('C'))
    assertEqual(t, raw[3], byte('P'))

    headerSize := raw[4:8]
    assertEqual(t, end.Uint32(headerSize), uint32(16))

    msgSize := raw[8:12]
    assertEqual(t, end.Uint32(msgSize), uint32(9))  // !1PWR01 = 7 chars

    // version
    assertEqual(t, raw[12], byte(0x01))

    // payload
    payload := raw[16:]
    assertEqual(t, payload, []byte("!1PWR01\r\n"))
}
