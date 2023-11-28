package main

import (
	"flag"
	"fmt"
	"net"
)

func serveRequest(req Message, resolver *net.UDPAddr) (Message, error) {
	resolverConn, err := net.DialUDP("udp", nil, resolver)
	if err != nil {
		return Message{}, fmt.Errorf("can't dial resolver: %w", err)
	}
	defer resolverConn.Close()

	answers := []Answer{}
	buf := make([]byte, 512)

	for _, q := range req.Questions {
		msg := Message{
			Header:    req.Header,
			Questions: []Question{q},
		}
		msg.Header.QDCount = 1
		msg.Header.ANCount = 0

		encoded, err := msg.Encode()
		if err != nil {
			return Message{}, fmt.Errorf("can't encode message to resolver: %w", err)
		}

		if _, err := resolverConn.Write(encoded); err != nil {
			return Message{}, fmt.Errorf("can't send message to resolver: %w", err)
		}

		n, err := resolverConn.Read(buf)
		if err != nil {
			return Message{}, fmt.Errorf("can't read response from resolver: %w", err)
		}

		resolverResp, err := DecodeMessage(buf[:n])
		if err != nil {
			return Message{}, fmt.Errorf("can't decode message from resolver: %w", err)
		}

		answers = append(answers, resolverResp.Answers...)
	}

	resp := Message{
		Header:    req.Header,
		Questions: req.Questions,
		Answers:   answers,
	}
	resp.Header.ANCount = uint16(len(answers))
	resp.Header.SetQR(true)
	if resp.Header.OpCode() != 0 {
		resp.Header.SetRC(4)
	}

	return resp, nil
}

func main() {
	var resolver string
	flag.StringVar(&resolver, "resolver", "", "Address where packet should be forwareded to")
	flag.Parse()

	if resolver == "" {
		fmt.Println("Resolver is not set")
		flag.Usage()
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	resolverAddr, err := net.ResolveUDPAddr("udp", resolver)
	if err != nil {
		fmt.Println("Failed to resolve resolver's UDP address:", err)
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

		receivedData := buf[:size]
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		req, err := DecodeMessage([]byte(receivedData))
		if err != nil {
			fmt.Printf("Error decoding packet: %v\n", err)
			return
		}

		resp, err := serveRequest(req, resolverAddr)
		if err != nil {
			fmt.Printf("Failed to serve request: %v\n", err)
			return
		}

		encoded, err := resp.Encode()
		if err != nil {
			fmt.Println("Failed to encode response: ", err)
			return
		}

		_, err = udpConn.WriteToUDP(encoded, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
