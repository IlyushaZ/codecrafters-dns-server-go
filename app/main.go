package main

import (
	"fmt"
	"net"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		req, err := DecodeMessage(buf)
		if err != nil {
			fmt.Printf("Error decoding packet: %v\n", err)
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		resp := Message{
			Header: Header{
				ID:      req.Header.ID,
				QDCount: req.Header.QDCount,
				ANCount: req.Header.QDCount, // TODO: add actual answer count when logics is implemented
				Flags:   req.Header.Flags,
			},
			Questions: make([]Question, 0, len(req.Questions)),
			Answers:   make([]Answer, 0, len(req.Questions)),
		}
		for _, q := range req.Questions {
			resp.Questions = append(resp.Questions, Question{
				Name:         q.Name,
				QuestionType: q.QuestionType,
				Class:        q.Class,
			})

			resp.Answers = append(resp.Answers, Answer{
				Name:       q.Name,
				RecordType: 1,
				Class:      1,
				TTL:        60,
				Length:     IPv4Len,
				Data:       3221225000, // example ip addr
			})
		}
		resp.Header.SetQR(true)
		if req.Header.OpCode() != 0 {
			resp.Header.SetRC(4)
		}

		encoded, err := resp.Encode()
		if err != nil {
			fmt.Println("Failed to encode response: ", err)
		}

		_, err = udpConn.WriteToUDP(encoded, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
