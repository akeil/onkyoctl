package onkyoctl

import (
	"encoding/binary"
	"testing"
)

func TestISCPFormat(t *testing.T) {
	command := ISCPCommand("PWR01")
	msg := NewISCPMessage(command)
	s := msg.Format()
	assertEqual(t, s, "!1PWR01\r\n")
}

func TestISCPParse(t *testing.T) {

	type Case struct {
		Data        []byte
		ExpectError bool
		Command     ISCPCommand
	}
	cases := []Case{
		// messages too short
		{
			Data:        make([]byte, 2),
			ExpectError: true,
		},
		{
			Data:        make([]byte, 9),
			ExpectError: true,
		},
		{
			Data:        []byte("1PWR01\n"),
			ExpectError: true,
		},
		{
			Data:        []byte("!PWR01\n"),
			ExpectError: true,
		},
		// various end styles
		{
			Data:        []byte("!1PWR01\r\n"),
			ExpectError: false,
			Command:     "PWR01",
		},
		{
			Data:        []byte("!1PWR01\r"),
			ExpectError: false,
			Command:     "PWR01",
		},
		{
			Data:        []byte("!1PWR01\n"),
			ExpectError: false,
			Command:     "PWR01",
		},
		// no end marker (invalid according to spec, but we accept it)
		{
			Data:        []byte("!1PWR01"),
			ExpectError: false,
			Command:     "PWR01",
		},
		//invalid end styles
		/*
		   Case{
		       Data: []byte("!1PWR01\n\n"),
		       ExpectError: true,
		   },
		   Case{
		       Data: []byte("!1PWR01\r\r"),
		       ExpectError: true,
		   },
		*/
	}
	for _, testCase := range cases {

		iscp, err := ParseISCP(testCase.Data)
		if testCase.ExpectError {
			assertErr(t, err)
		} else {
			assertNoErr(t, err)
			assertEqual(t, iscp.Command(), testCase.Command)
		}
	}
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
	assertEqual(t, end.Uint32(msgSize), uint32(9)) // !1PWR01 = 7 chars

	// version
	assertEqual(t, raw[12], byte(0x01))

	// payload
	payload := raw[16:]
	assertEqual(t, payload, []byte("!1PWR01\r\n"))
}

func TestEISCPParse(t *testing.T) {
	m := NewEISCPMessage("PWR01")
	raw := m.Raw()
	eiscp, err := ParseEISCP(raw)
	assertNoErr(t, err)
	assertEqual(t, eiscp.Command(), m.Command())

	_, err = ParseEISCP(make([]byte, 1))
	assertErr(t, err)

	_, err = ParseEISCP(make([]byte, 100))
	assertErr(t, err)

	// magic ok, message lengths bad
	_, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	})
	assertErr(t, err)

	// magic ok, header length ok, missing part of header
	_, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x00,
	})
	assertErr(t, err)

	// magic ok, header length ok, missing payload
	_, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, // version 1, 3x reserved
	})
	assertErr(t, err)

	// magic ok, header length ok, missing payload length
	_, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x00, // 0
		0x01, 0x00, 0x00, 0x00, // version 1, 3x reserved
		0x21, 0x31, 0x58, 0x58, 0x0A, // !1XX\n
	})
	assertErr(t, err)

	// magic ok, header length ok, invalid payload length
	_, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x0C, // 12
		0x01, 0x00, 0x00, 0x00, // version 1, 3x reserved
		0x21, 0x31, 0x58, 0x58, 0x0A, // !1XX\n
	})
	assertErr(t, err)

	// valid message, w/o EOF
	eiscp, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x05, // 5
		0x01, 0x00, 0x00, 0x00, // version 1, 3x reserved
		0x21, 0x31, 0x58, 0x58, 0x58, 0x0A, // !1XXX\n
	})
	assertNoErr(t, err)
	assertEqual(t, eiscp.Command(), ISCPCommand("XXX"))

	// valid message, w/ EOF
	eiscp, err = ParseEISCP([]byte{
		0x49, 0x53, 0x43, 0x50, // ISCP
		0x00, 0x00, 0x00, 0x10, // 16
		0x00, 0x00, 0x00, 0x05, // 5
		0x01, 0x00, 0x00, 0x00, // version 1, 3x reserved
		0x21, 0x31, 0x58, 0x58, 0x58, 0x0A, // !1XXX\n
		0x1A, // EOF
	})
	assertNoErr(t, err)
	assertEqual(t, eiscp.Command(), ISCPCommand("XXX"))
}
