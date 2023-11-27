package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const IPv4Len = 4

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

type Label struct {
	Len     byte
	Content []byte
}

type Question struct {
	Name         []Label
	QuestionType uint16
	Class        uint16
}

type Answer struct {
	Name       []Label
	RecordType uint16
	Class      uint16
	TTL        uint32
	Length     uint16
	Data       uint32 // ATM we only support class "A" record type which means that data will only contain IPv4
}

type Message struct {
	Header    Header
	Questions []Question
	Answers   []Answer
}

func (h *Header) SetQR(val bool) {
	if val {
		h.Flags |= 1 << 15
	}
}

func (h *Header) SetRC(val uint16) {
	h.Flags |= val
}

func (h *Header) OpCode() uint16 {
	return (h.Flags & 0x7800) >> 11
}

func (m *Message) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}

	if err := binary.Write(buf, binary.BigEndian, m.Header); err != nil {
		return nil, fmt.Errorf("can't write header: %w", err)
	}

	for _, q := range m.Questions {
		for _, l := range q.Name {
			if err := buf.WriteByte(l.Len); err != nil {
				return nil, fmt.Errorf("can't write label's len in question: %w", err)
			}
			if _, err := buf.Write(l.Content); err != nil {
				return nil, fmt.Errorf("can't write label's content in question: %w", err)
			}
		}
		if err := buf.WriteByte('\x00'); err != nil {
			return nil, fmt.Errorf("can't write terminating byte of name in question: %w", err)
		}

		typeAndClass := make([]byte, 0, 4)
		typeAndClass = binary.BigEndian.AppendUint16(typeAndClass, q.QuestionType)
		typeAndClass = binary.BigEndian.AppendUint16(typeAndClass, q.Class)

		if _, err := buf.Write(typeAndClass); err != nil {
			return nil, fmt.Errorf("can't write type and class: %w", err)
		}
	}

	for _, a := range m.Answers {
		for _, l := range a.Name {
			if err := buf.WriteByte(l.Len); err != nil {
				return nil, fmt.Errorf("can't write label's len in answer: %w", err)
			}
			if _, err := buf.Write(l.Content); err != nil {
				return nil, fmt.Errorf("can't write label's content in answer: %w", err)
			}
		}
		if err := buf.WriteByte('\x00'); err != nil {
			return nil, fmt.Errorf("can't write terminating byte of name in answer: %w", err)
		}

		// restAnswer is the remaining part of answer, excluding name
		restAnswer := make([]byte, 0, 14)
		restAnswer = binary.BigEndian.AppendUint16(restAnswer, a.RecordType)
		restAnswer = binary.BigEndian.AppendUint16(restAnswer, a.Class)
		restAnswer = binary.BigEndian.AppendUint32(restAnswer, a.TTL)
		restAnswer = binary.BigEndian.AppendUint16(restAnswer, a.Length)
		restAnswer = binary.BigEndian.AppendUint32(restAnswer, a.Data)

		if _, err := buf.Write(restAnswer); err != nil {
			return nil, fmt.Errorf("can't write answer: %w", err)
		}
	}

	return buf.Bytes(), nil
}

func DecodeMessage(packet []byte) (Message, error) {
	h := Header{}

	if len(packet) < 12 {
		return Message{}, errors.New("malformed packet")
	}

	h.ID = binary.BigEndian.Uint16(packet[:2])
	h.Flags = binary.BigEndian.Uint16(packet[2:4])
	h.QDCount = binary.BigEndian.Uint16(packet[4:6])
	h.ANCount = binary.BigEndian.Uint16(packet[6:8])
	h.NSCount = binary.BigEndian.Uint16(packet[8:10])
	h.ARCount = binary.BigEndian.Uint16(packet[10:12])

	buf := bytes.NewBuffer(packet[12:])
	qs := make([]Question, 0, h.QDCount)

	for i := 0; i < int(h.QDCount); i++ {
		q := Question{}

	NameLoop:
		for {
			labelLen, err := buf.ReadByte()
			if err != nil {
				return Message{}, fmt.Errorf("can't read label's len: %w", err)
			}

			if labelLen == '\x00' {
				break NameLoop
			}

			labelContent := make([]byte, labelLen)
			read, err := buf.Read(labelContent)
			if err != nil {
				return Message{}, fmt.Errorf("can't read label's content: %w", err)
			}
			if read != int(labelLen) {
				return Message{}, fmt.Errorf("malformed label. expected len to be %d, got %d", labelLen, read)
			}

			q.Name = append(q.Name, Label{
				Len:     labelLen,
				Content: labelContent,
			})
		}

		restQuestion := make([]byte, 4)
		read, err := buf.Read(restQuestion)
		if err != nil {
			return Message{}, fmt.Errorf("can't read question's type and class: %w", err)
		}
		if read != 4 {
			return Message{}, errors.New("malformed question")
		}

		q.QuestionType = binary.BigEndian.Uint16(restQuestion[:2])
		q.Class = binary.BigEndian.Uint16(restQuestion[2:])

		qs = append(qs, q)
	}

	return Message{Header: h, Questions: qs}, nil
}
