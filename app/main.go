package main

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/dns-server-starter-go/app/dns"
)

// Ensures gofmt doesn't remove the "net" import in stage 1 (feel free to remove this!)
var _ = net.ListenUDP

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	
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
	
		// Create an empty response
		dnsMessage := dns.DNSMessage{}
		dnsMessage.Header = dns.Header{
			ID: 1234,
			QR: 1,
			OpCode: 0,
			AA: 0,
			TC: 0,
			RD: 0,
			RA: 0,
			Z: 0,
			RCode: 0,
			QDCount: 1,
			ANCount: 0,
			NSCount: 0,
			ARCount: 0,
		}
		dnsMessage.Question = dns.Question{
			Question: "codecrafters.io",
			Type: 1,
			Class: 1,
			
		}
		dnsMessage.Answer = dns.Answer{
			Domain: "codecrafters.io",
			Type: 1,
			Class: 1,
			TTL:60,
			Len:4,
			Data:"8.8.8.8",
		}
		response := dnsMessage.ParseMsg();
		
		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
