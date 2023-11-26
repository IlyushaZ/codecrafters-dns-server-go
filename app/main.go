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
				QDCount: 1,
				ANCount: 1,
				Flags:   req.Header.Flags,
			},
			Question: Question{
				Name: []Label{
					{Len: 12, Content: []byte("codecrafters")},
					{Len: 2, Content: []byte("io")},
				},
				QuestionType: 1,
				Class:        1,
			},
			Answer: Answer{
				Name: []Label{
					{Len: 12, Content: []byte("codecrafters")},
					{Len: 2, Content: []byte("io")},
				},
				RecordType: 1,
				Class:      1,
				TTL:        60,
				Length:     IPv4Len,
				Data:       3221225000, // example ip address
			},
		}
		resp.Header.SetQR(true)
		resp.Header.SetRC()

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
