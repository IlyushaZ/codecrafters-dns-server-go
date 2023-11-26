package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Header struct {
	ID uint16
	// Flags contains multiple values:
	// QR (1 bit) - query/response.
	// OPCODE (4 bits) - operation code.
	// AA (1 bit) - if server "owns" queried domain.
	// TC (1 bit) - if message is larger than 512 bytes. Can be used as indicator of whether TCP should be used.
	// RD (1 bit) - if recursion is desired. Set by client.
	// RA (1 bit) - if recursion is available. Set by server.
	// Z (3 bits) - reserved.
	// RC (4 bits) - response code and reason in case if request is failed.
	Flags uint16
	// QDCount is number of entries in the question section
	QDCount uint16
	// ANCount is number of entries in the answer section
	ANCount uint16
	// NSCount is number of entries in the authority section
	NSCount uint16
	// ARCound is number of entries in the additional section
	ARCount uint16
}

type Message struct {
	Header Header
}

func (h *Header) SetQR(val bool) {
	if val {
		h.Flags |= 1 << 15
	}
}

func (m *Message) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}

	if err := binary.Write(buf, binary.BigEndian, m); err != nil {
		return nil, fmt.Errorf("can't write to buffer: %w", err)
	}

	return buf.Bytes(), nil
}
