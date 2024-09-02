package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/dns-server-starter-go/app/dns"
)

// Ensures gofmt doesn't remove the "net" import in stage 1 (feel free to remove this!)
var _ = net.ListenUDP

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	Resolver := flag.String("resolver", "8.8.8.8:80", "resolver directory")
	flag.Parse()
	fmt.Println("resolver - ", *Resolver)
	resolverUdp, err := net.ResolveUDPAddr("udp", *Resolver)
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	if err!=nil{
		panic("Could not connect to resolver server")
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
		fmt.Println()
		dnsQuery := dns.ParseDNSMessage(buf);
		// Create an empty response
		dnsMessage := dns.DNSMessage{}
		dnsMessage.Answer = []dns.Answer{
		}
		for i:=0;i<len(dnsQuery.Question);i++{
			dnsQueryCopy := dnsQuery;
			dnsQueryCopy.Question = dnsQuery.Question[i:i+1]
			dnsQueryCopy.Header.QDCount = 1;
			dnsQueryCopy.Header.ANCount = 0;
			dnsQueryCopy.Header.ID = uint16(int(dnsQueryCopy.Header.ID) + 1)
			fmt.Println(len(dnsQueryCopy.Question))
			req := dnsQueryCopy.ParseMsg()
			conn, err := net.DialUDP("udp", nil, resolverUdp)
			_, err = conn.Write(req);
			if(err!=nil){
				fmt.Println("Error sending:", err)
				return
			}
			res := make([]byte, 512)
			_ , err = conn.Read(res)
			if err != nil {
				fmt.Println("Error receiving:", err)
				return
			}
			resVal := dns.ParseDNSMessage(res);
			fmt.Println(resVal.Header.QDCount, " - ", resVal.Header.ANCount)
			if(len(resVal.Answer)>0){
				fmt.Println(resVal.Answer[0].Domain, resVal.Answer[0].Data)
			}else{
				fmt.Println("empty")
			}
			dnsMessage.Answer = append(dnsMessage.Answer, resVal.Answer...)
			conn.Close()
		}

		rcode:=0;
		if(dnsQuery.Header.OpCode!=0){
			rcode=4;
		}
		
		dnsMessage.Question = []dns.Question{
		}
		for i:=0;i<int(dnsQuery.Header.QDCount);i++{
			dnsMessage.Question = append(dnsMessage.Question, dns.Question{
				Question: dnsQuery.Question[i].Question,
				Type: 1,
				Class: 1,
			},)	
		}
		dnsMessage.Header = dns.Header{
			ID: dnsQuery.Header.ID,
			QR: 1,
			OpCode: dnsQuery.Header.OpCode,
			AA: 0,
			TC: 0,
			RD: dnsQuery.Header.RD,
			RA: 0,
			Z: 0,
			RCode: uint8(rcode),
			QDCount: uint16(len(dnsMessage.Question)),
			ANCount: uint16(len(dnsMessage.Answer)),
			NSCount: 0,
			ARCount: 0,
		}
		// for i:=0;i<len(dnsQuery.Question);i++{
		// 	fmt.Println(dnsQuery.Question[i].Question);
		// 	dnsMessage.Answer = append(dnsMessage.Answer, dns.Answer{
		// 		Domain: dnsQuery.Question[i].Question,
		// 		Type: 1,
		// 		Class: 1,
		// 		TTL:60,
		// 		Len:4,
		// 		Data:"8.8.8.8",},)	
		// }
		response := dnsMessage.ParseMsg();

		
		_, err = udpConn.WriteToUDP(response, source)
		time.Sleep(500*time.Millisecond)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
