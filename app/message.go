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

type Question struct {
	Name         []Label
	QuestionType uint16
	Class        uint16
}

type Label struct {
	Len     byte
	Content []byte
}

type Message struct {
	Header   Header
	Question Question
}

func (h *Header) SetQR(val bool) {
	if val {
		h.Flags |= 1 << 15
	}
}

func (m *Message) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}

	if err := binary.Write(buf, binary.BigEndian, m.Header); err != nil {
		return nil, fmt.Errorf("can't write header: %w", err)
	}

	for _, l := range m.Question.Name {
		if err := buf.WriteByte(l.Len); err != nil {
			return nil, fmt.Errorf("can't write label's len: %w", err)
		}
		if _, err := buf.Write(l.Content); err != nil {
			return nil, fmt.Errorf("can't write label's content: %w", err)
		}
	}
	if err := buf.WriteByte('\x00'); err != nil {
		return nil, fmt.Errorf("can't write terminating byte of name: %w", err)
	}

	typeAndClass := make([]byte, 0, 32)
	typeAndClass = binary.BigEndian.AppendUint16(typeAndClass, m.Question.QuestionType)
	typeAndClass = binary.BigEndian.AppendUint16(typeAndClass, m.Question.Class)

	if _, err := buf.Write(typeAndClass); err != nil {
		return nil, fmt.Errorf("can't write type and class: %w", err)
	}

	return buf.Bytes(), nil
}
