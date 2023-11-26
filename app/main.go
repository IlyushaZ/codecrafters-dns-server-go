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

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		msg := Message{
			Header: Header{
				ID:      1234, // TODO: set this value as ID from request
				QDCount: 1,
			},
			Question: Question{
				Name: []Label{
					{Len: 12, Content: []byte("codecrafters")},
					{Len: 2, Content: []byte("io")},
				},
				QuestionType: 1,
				Class:        1,
			},
		}
		msg.Header.SetQR(true)

		resp, err := msg.Encode()
		if err != nil {
			fmt.Println("Failed to encode response: ", err)
		}

		_, err = udpConn.WriteToUDP(resp, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
